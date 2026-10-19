package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxFile = 100 << 20

type App struct {
	H         *apphost.Host
	mu        sync.Mutex
	storageMu sync.Mutex
	order     []string
	pending   map[string]Request
	cancel    map[string]context.CancelFunc
	wake      chan struct{}
}
type Document struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Hash    string    `json:"-"`
	Size    int64     `json:"size"`
	Created string    `json:"created"`
	Drive   DriveSave `json:"drive"`
}
type Request struct {
	Operation   string   `json:"operation"`
	Documents   []string `json:"documents"`
	Pages       []string `json:"pages"`
	Rotation    int      `json:"rotation"`
	Password    string   `json:"password,omitempty"`
	SaveToDrive bool     `json:"saveToDrive,omitempty"`
}
type Job struct {
	ID            string          `json:"id"`
	Operation     string          `json:"operation"`
	QueuePosition int             `json:"queuePosition"`
	State         string          `json:"state"`
	Message       string          `json:"message"`
	Result        string          `json:"result"`
	Created       string          `json:"created"`
	Finished      string          `json:"finished"`
	Metadata      json.RawMessage `json:"metadata"`
}

func New(h *apphost.Host) (*App, error) {
	var schemaVersion int
	if e := h.DB.QueryRow("SELECT coalesce(max(version),0) FROM host_migrations").Scan(&schemaVersion); e != nil {
		return nil, e
	}
	if schemaVersion < 1 {
		if e := h.Migrate(1, `CREATE TABLE pdf_documents(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,name TEXT NOT NULL,hash TEXT NOT NULL,size INTEGER NOT NULL,created TEXT NOT NULL); CREATE INDEX pdf_owner ON pdf_documents(user_id); CREATE TABLE pdf_annotations(document_id TEXT PRIMARY KEY REFERENCES pdf_documents(id) ON DELETE CASCADE,value TEXT NOT NULL); CREATE TABLE pdf_artifacts(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,hash TEXT NOT NULL,size INTEGER NOT NULL); CREATE TABLE pdf_jobs(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,operation TEXT NOT NULL,state TEXT NOT NULL,message TEXT NOT NULL,result TEXT NOT NULL DEFAULT '',created TEXT NOT NULL,finished TEXT NOT NULL DEFAULT '',metadata TEXT NOT NULL DEFAULT '{}');`); e != nil {
			return nil, e
		}
	}
	if e := h.Migrate(2, `CREATE TABLE pdf_drive(document_id TEXT PRIMARY KEY REFERENCES pdf_documents(id) ON DELETE CASCADE,state TEXT NOT NULL,source TEXT NOT NULL DEFAULT '',path TEXT NOT NULL DEFAULT '',message TEXT NOT NULL DEFAULT '',lease TEXT NOT NULL DEFAULT '',expires INTEGER NOT NULL DEFAULT 0);`); e != nil {
		return nil, e
	}
	h.DB.Exec("UPDATE pdf_jobs SET state='cancelled',message='Interrupted by restart; original files are unchanged' WHERE state IN('queued','processing')")
	a := &App{H: h, pending: map[string]Request{}, cancel: map[string]context.CancelFunc{}, wake: make(chan struct{}, 1)}
	a.cleanJobs()
	a.gcFiles()
	m := h.Mux
	m.HandleFunc("GET /api/documents", h.Require(a.documents))
	m.HandleFunc("POST /api/documents", h.Require(a.upload))
	m.HandleFunc("GET /api/documents/{id}", h.Require(a.download))
	m.HandleFunc("GET /api/documents/{id}/annotations", h.Require(a.annotations))
	m.HandleFunc("PUT /api/documents/{id}/annotations", h.Require(a.annotations))
	m.HandleFunc("DELETE /api/documents/{id}", h.Require(a.delete))
	m.HandleFunc("POST /api/documents/{id}/drive-save", h.Require(a.driveSave))
	m.HandleFunc("POST /api/jobs", h.Require(a.createJob))
	m.HandleFunc("DELETE /api/artifacts/{id}", h.Require(a.deleteArtifact))
	m.HandleFunc("GET /api/artifacts/{id}", h.Require(a.artifact))
	m.HandleFunc("GET /api/jobs", h.Require(a.jobs))
	m.HandleFunc("DELETE /api/jobs/{id}", h.Require(a.cancelJob))
	return a, nil
}
func (a *App) doc(user, id string) (d Document, e error) {
	e = a.H.DB.QueryRow("SELECT id,name,hash,size,created FROM pdf_documents WHERE id=? AND user_id=?", id, user).Scan(&d.ID, &d.Name, &d.Hash, &d.Size, &d.Created)
	return
}
func (a *App) documents(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT d.id,d.name,d.size,d.created,COALESCE(s.state,'local'),COALESCE(s.source,''),COALESCE(s.path,''),COALESCE(s.message,'') FROM pdf_documents d LEFT JOIN pdf_drive s ON s.document_id=d.id WHERE d.user_id=? ORDER BY d.created DESC LIMIT 500", apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Document{}
	for rows.Next() {
		var d Document
		if e = rows.Scan(&d.ID, &d.Name, &d.Size, &d.Created, &d.Drive.State, &d.Drive.Source, &d.Drive.Path, &d.Drive.Message); e != nil {
			a.H.DBError(w, e)
			return
		}
		out = append(out, d)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	control := http.NewResponseController(w)
	_ = control.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer control.SetReadDeadline(time.Time{})
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	name := strings.TrimSpace(r.Header.Get("X-File-Name"))
	if name == "" || len(name) > 200 || strings.ContainsAny(name, "/\\\r\n\x00") || !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		apphost.Problem(w, 400, "Choose a PDF with a short filename")
		return
	}
	var count int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM pdf_documents WHERE user_id=?", apphost.Current(r).User.ID).Scan(&count)
	if count >= 500 {
		apphost.Problem(w, 413, "Remove old documents before uploading more")
		return
	}
	head := make([]byte, 5)
	if _, e := io.ReadFull(r.Body, head); e != nil || string(head) != "%PDF-" {
		apphost.Problem(w, 400, "This file does not have a PDF header")
		return
	}
	budget := a.budget(apphost.Current(r).User.ID)
	if budget < 1 {
		apphost.Problem(w, 413, "My PDF storage limit is 1 GiB per account. Remove old documents or split archives first")
		return
	}
	hash, size, e := a.H.PutBlob(io.MultiReader(bytes.NewReader(head), r.Body), min(int64(maxFile), budget))
	if e != nil {
		apphost.Problem(w, 413, "PDF upload failed; the limit is 100 MiB")
		return
	}
	d := Document{ID: apphost.ID(), Name: name, Hash: hash, Size: size, Created: time.Now().UTC().Format(time.RFC3339Nano)}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer tx.Rollback()
	_, e = tx.Exec("INSERT INTO pdf_documents VALUES(?,?,?,?,?,?)", d.ID, apphost.Current(r).User.ID, d.Name, d.Hash, d.Size, d.Created)
	if e == nil && r.Header.Get("X-Save-To-Drive") == "true" {
		_, e = tx.Exec("INSERT INTO pdf_drive(document_id,state) VALUES(?,'pending')", d.ID)
		d.Drive.State = "pending"
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, d)
}
func (a *App) download(w http.ResponseWriter, r *http.Request) {
	d, e := a.doc(apphost.Current(r).User.ID, r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, "Document not found")
		return
	}
	path, e := a.H.BlobPath(d.Hash)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="document.pdf"`)
	http.ServeFile(w, r, path)
}
func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	_, e := a.H.DB.Exec("DELETE FROM pdf_documents WHERE id=? AND user_id=?", r.PathValue("id"), apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.DB.Exec("UPDATE pdf_jobs SET result='',message='Result removed; originals are unchanged' WHERE result=? AND user_id=?", r.PathValue("id"), apphost.Current(r).User.ID)
	a.gcFiles()
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func validRequest(in Request) error {
	if len(in.Documents) == 0 || len(in.Documents) > 10 || len(in.Password) > 256 || len(in.Pages) > 1000 {
		return errors.New("Choose up to ten PDFs and a bounded page selection")
	}
	switch in.Operation {
	case "inspect", "split", "merge", "extract", "remove", "rotate", "compress", "metadata-remove", "blank":
	default:
		return errors.New("Unsupported PDF operation")
	}
	if in.Operation != "merge" && len(in.Documents) != 1 {
		return errors.New("This operation takes one PDF")
	}
	if in.Operation == "rotate" && (in.Rotation == 0 || in.Rotation%90 != 0 || in.Rotation < -270 || in.Rotation > 270) {
		return errors.New("Choose a rotation of 90, 180 or 270 degrees")
	}
	for _, p := range in.Pages {
		if len(p) > 30 || strings.ContainsAny(p, "/\\\r\n") {
			return errors.New("Invalid page selection")
		}
	}
	return nil
}
func (a *App) createJob(w http.ResponseWriter, r *http.Request) {
	var in Request
	if !apphost.Decode(w, r, &in) {
		return
	}
	if e := validRequest(in); e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	u := apphost.Current(r).User.ID
	for _, id := range in.Documents {
		if _, e := a.doc(u, id); e != nil {
			apphost.Problem(w, 404, "Document not found")
			return
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.pending) >= 20 {
		apphost.Problem(w, 429, "The PDF queue is full; wait for a job to finish")
		return
	}
	id := apphost.ID()
	_, e := a.H.DB.Exec("INSERT INTO pdf_jobs(id,user_id,operation,state,message,created)VALUES(?,?,?,'queued','Waiting for the bounded PDF worker',?)", id, u, in.Operation, time.Now().UTC().Format(time.RFC3339Nano))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	a.pending[id] = in
	a.order = append(a.order, id)
	select {
	case a.wake <- struct{}{}:
	default:
	}
	apphost.JSON(w, 202, map[string]string{"id": id, "state": "queued"})
}
func (a *App) jobs(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	positions := map[string]int{}
	position := 0
	for _, id := range a.order {
		if _, ok := a.pending[id]; ok {
			position++
			positions[id] = position
		}
	}
	a.mu.Unlock()
	rows, e := a.H.DB.Query("SELECT id,operation,state,message,result,created,finished,metadata FROM pdf_jobs WHERE user_id=? ORDER BY created DESC LIMIT 100", apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Job{}
	for rows.Next() {
		var j Job
		var metadata string
		if e = rows.Scan(&j.ID, &j.Operation, &j.State, &j.Message, &j.Result, &j.Created, &j.Finished, &metadata); e != nil {
			a.H.DBError(w, e)
			return
		}
		j.Metadata = json.RawMessage(metadata)
		j.QueuePosition = positions[j.ID]
		out = append(out, j)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) cancelJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var owner string
	if a.H.DB.QueryRow("SELECT user_id FROM pdf_jobs WHERE id=?", id).Scan(&owner) != nil || owner != apphost.Current(r).User.ID {
		apphost.Problem(w, 404, "Job not found")
		return
	}
	a.mu.Lock()
	delete(a.pending, id)
	if cancel := a.cancel[id]; cancel != nil {
		cancel()
	}
	a.mu.Unlock()
	a.H.DB.Exec("UPDATE pdf_jobs SET state='cancelled',message='Cancelled; original files are unchanged',finished=? WHERE id=? AND state IN('queued','processing')", time.Now().UTC().Format(time.RFC3339Nano), id)
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) Run(ctx context.Context) error {
	defer func() {
		a.mu.Lock()
		for id := range a.pending {
			a.H.DB.Exec("UPDATE pdf_jobs SET state='cancelled',message='Server stopped; retry this operation' WHERE id=?", id)
		}
		a.mu.Unlock()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-a.wake:
		}
		for {
			a.mu.Lock()
			id := ""
			var in Request
			for len(a.order) > 0 {
				k := a.order[0]
				a.order = a.order[1:]
				v, ok := a.pending[k]
				if !ok {
					continue
				}
				id = k
				in = v
				delete(a.pending, k)
				break
			}
			a.mu.Unlock()
			if id == "" {
				break
			}
			a.H.Freeze.RLock()
			a.process(ctx, id, in)
			a.H.Freeze.RUnlock()
			if ctx.Err() != nil {
				return nil
			}
		}
	}
}
func (a *App) process(parent context.Context, id string, in Request) {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	a.mu.Lock()
	a.cancel[id] = cancel
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.cancel, id); a.mu.Unlock() }()
	var owner, state string
	if a.H.DB.QueryRow("SELECT user_id,state FROM pdf_jobs WHERE id=?", id).Scan(&owner, &state) != nil || state != "queued" {
		return
	}
	fail := func(message string) {
		a.H.DB.Exec("UPDATE pdf_jobs SET state='failed',message=?,finished=? WHERE id=? AND state<>'cancelled'", message, time.Now().UTC().Format(time.RFC3339Nano), id)
	}
	dir, e := os.MkdirTemp(a.H.Config.Data, "pdf-work-")
	if e != nil {
		fail("Could not prepare a protected job folder")
		return
	}
	defer os.RemoveAll(dir)
	wr := WorkerRequest{Operation: in.Operation, Pages: in.Pages, Rotation: in.Rotation, Password: in.Password, Output: filepath.Join(dir, "result.pdf")}
	for i, docID := range in.Documents {
		d, e := a.doc(owner, docID)
		if e != nil {
			fail("A source document was removed")
			return
		}
		path, e := a.H.BlobPath(d.Hash)
		if e != nil {
			fail("Source file is unavailable")
			return
		}
		source, e := os.Open(path)
		if e != nil {
			fail("Source file is unavailable")
			return
		}
		dest := filepath.Join(dir, "input-"+string(rune('a'+i))+".pdf")
		f, e := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e == nil {
			_, e = io.Copy(f, source)
			f.Close()
		}
		source.Close()
		if e != nil {
			fail("Could not prepare source file")
			return
		}
		wr.Inputs = append(wr.Inputs, dest)
	}
	a.H.DB.Exec("UPDATE pdf_jobs SET state='processing',message='Validating and processing PDF; up to 90 seconds' WHERE id=? AND state='queued'", id)
	payload, _ := json.Marshal(wr)
	var out limitedBuffer
	exe, _ := os.Executable()
	e = apphost.Worker(ctx, exe, dir, bytes.NewReader(payload), &out)
	if e != nil {
		if ctx.Err() != nil {
			fail("Job stopped or exceeded its 90-second limit; original files are unchanged")
		} else {
			fail("PDF could not be processed. Check its password/pages or try a smaller file; workers are limited to 512 MiB and 60 CPU seconds. Original is unchanged.")
		}
		return
	}
	var metadata map[string]any
	if json.Unmarshal(out.Bytes(), &metadata) != nil {
		fail("Worker returned an invalid result")
		return
	}
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	defer a.gcFiles()
	result := ""
	var hash string
	var size int64
	if in.Operation == "split" {
		wr.Output = filepath.Join(dir, "split-pages.zip")
		if e = zipPages(filepath.Join(dir, "split"), wr.Output); e != nil {
			fail("Split output exceeds its size or page-count limit")
			return
		}
	}
	if in.Operation != "inspect" {
		f, e := os.Open(wr.Output)
		if e != nil {
			fail("Worker did not produce its output")
			return
		}
		hash, size, e = a.H.PutBlob(f, min(int64(maxFile), a.budget(owner)))
		f.Close()
		if e != nil {
			fail("Result exceeds 100 MiB or your 1 GiB storage quota; remove old documents and retry")
			return
		}
		result = apphost.ID()
	}
	if ctx.Err() != nil {
		fail("Job stopped before saving; original files are unchanged")
		return
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		fail("Could not save the job result")
		return
	}
	defer tx.Rollback()
	if e = tx.QueryRow("SELECT state FROM pdf_jobs WHERE id=?", id).Scan(&state); e != nil || state != "processing" {
		return
	}
	if result != "" {
		if in.Operation == "split" {
			_, e = tx.Exec("INSERT INTO pdf_artifacts VALUES(?,?,?,?)", result, owner, hash, size)
		} else {
			_, e = tx.Exec("INSERT INTO pdf_documents VALUES(?,?,?,?,?,?)", result, owner, "processed-"+in.Operation+".pdf", hash, size, time.Now().UTC().Format(time.RFC3339Nano))
			if e == nil && in.SaveToDrive {
				_, e = tx.Exec("INSERT INTO pdf_drive(document_id,state) VALUES(?,'pending')", result)
			}
		}
		if e != nil {
			tx.Rollback()
			fail("Could not save the result")
			return
		}
	}
	if _, e = tx.Exec("UPDATE pdf_jobs SET state='complete',message='Complete; originals are unchanged',result=?,metadata=?,finished=? WHERE id=? AND state='processing'", result, string(out.Bytes()), time.Now().UTC().Format(time.RFC3339Nano), id); e != nil {
		tx.Rollback()
		fail("Could not complete the job")
		return
	}
	if e = tx.Commit(); e != nil {
		fail("Could not commit the result; retry this job")
	}

}

var _ = sql.ErrNoRows
