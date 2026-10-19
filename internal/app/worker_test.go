package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == "--worker" {
		if Worker(os.Stdin, os.Stdout) != nil {
			os.Exit(2)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func fixturePDF(pages int) []byte {
	var b bytes.Buffer
	fmt.Fprint(&b, "%PDF-1.7\n")
	offset := []int{0}
	write := func(n int, s string) {
		offset = append(offset, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", n, s)
	}
	write(1, "<< /Type /Catalog /Pages 2 0 R >>")
	kids := ""
	for i := 0; i < pages; i++ {
		kids += fmt.Sprintf("%d 0 R ", 4+i*2)
	}
	write(2, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", kids, pages))
	write(3, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	for i := 0; i < pages; i++ {
		n := 4 + i*2
		write(n, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 400] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", n+1))
		content := "BT /F1 12 Tf 20 200 Td (PRIVATE-FIXTURE-SECRET) Tj ET"
		write(n+1, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offset))
	for _, o := range offset[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offset), xref)
	return b.Bytes()
}
func TestBoundedNativeWorkerPDFOperationsAndMalformation(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "input.pdf")
	if e := os.WriteFile(in, fixturePDF(3), 0600); e != nil {
		t.Fatal(e)
	}
	for _, op := range []string{"inspect", "rotate", "extract", "remove", "merge", "compress", "blank", "metadata-remove", "split"} {
		wr := WorkerRequest{Operation: op, Inputs: []string{in}, Output: filepath.Join(dir, op+".pdf"), Pages: []string{"1"}, Rotation: 90}
		if op == "merge" {
			wr.Inputs = append(wr.Inputs, in)
		}
		payload, _ := json.Marshal(wr)
		var out bytes.Buffer
		exe, _ := os.Executable()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		e := apphost.Worker(ctx, exe, dir, bytes.NewReader(payload), &out)
		cancel()
		if e != nil {
			t.Fatal(op, e)
		}
		if !strings.Contains(out.String(), `"pages":`) {
			t.Fatal("missing metadata", op)
		}
		if op == "split" {
			if e = zipPages(filepath.Join(dir, "split"), wr.Output); e != nil {
				t.Fatal(e)
			}
			continue
		}
		if op != "inspect" {
			f, e := os.Open(wr.Output)
			if e != nil {
				t.Fatal(e)
			}
			n, e := api.PageCount(context.Background(), f, model.NewStatelessConfiguration())
			f.Close()
			if e != nil || n < 1 {
				t.Fatal(op, n, e)
			}
		}
	}
	bad := filepath.Join(dir, "bad.pdf")
	os.WriteFile(bad, []byte("%PDF-broken"), 0600)
	wr := WorkerRequest{Operation: "inspect", Inputs: []string{bad}, Output: filepath.Join(dir, "badout.pdf")}
	payload, _ := json.Marshal(wr)
	exe, _ := os.Executable()
	if e := apphost.Worker(context.Background(), exe, dir, bytes.NewReader(payload), &bytes.Buffer{}); e == nil {
		t.Fatal("corrupted PDF accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := apphost.Worker(ctx, exe, dir, bytes.NewReader(payload), &bytes.Buffer{}); e == nil {
		t.Fatal("cancel ignored")
	}
	if raw, _ := os.ReadFile(in); !bytes.Equal(raw, fixturePDF(3)) {
		t.Fatal("source modified")
	}
}

type pdfClient struct {
	h      *apphost.Host
	cookie *http.Cookie
	csrf   string
}

func (c *pdfClient) call(method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, c.h.Config.Origin+path, bytes.NewReader(body))
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Origin", c.h.Config.Origin)
	r.Header.Set("X-CSRF-Token", c.csrf)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	if c.cookie != nil {
		r.AddCookie(c.cookie)
	}
	w := httptest.NewRecorder()
	c.h.Handler().ServeHTTP(w, r)
	return w
}
func pdfLogin(t *testing.T, c *pdfClient, email string) {
	b, _ := json.Marshal(map[string]string{"email": email, "password": "fixture-password"})
	w := c.call("POST", "/api/login", b, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	c.cookie = w.Result().Cookies()[0]
	w = c.call("GET", "/api/me", nil, nil)
	var s apphost.Session
	json.Unmarshal(w.Body.Bytes(), &s)
	c.csrf = s.CSRF
}
func TestDocumentIsolationTraversalAnnotationCSRFAndJobs(t *testing.T) {
	h, e := apphost.New(apphost.Config{ID: "pdf", Name: "My PDF", Data: filepath.Join(t.TempDir(), "data"), Origin: "http://localhost:29009", Listen: "127.0.0.1:29009"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	a, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	h.Bootstrap("owner@example.test", "fixture-password")
	o := &pdfClient{h: h}
	pdfLogin(t, o, "owner@example.test")
	w := o.call("POST", "/api/local-users", []byte(`{"email":"member@example.test","password":"fixture-password"}`), nil)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body)
	}
	m := &pdfClient{h: h}
	pdfLogin(t, m, "member@example.test")
	w = o.call("POST", "/api/documents", fixturePDF(2), map[string]string{"X-File-Name": "fixture.pdf"})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body)
	}
	var d Document
	json.Unmarshal(w.Body.Bytes(), &d)
	if m.call("GET", "/api/documents/"+d.ID, nil, nil).Code != 404 {
		t.Fatal("cross-user download")
	}
	if o.call("POST", "/api/documents", fixturePDF(1), map[string]string{"X-File-Name": "../escape.pdf"}).Code != 400 {
		t.Fatal("filename traversal")
	}
	if o.call("PUT", "/api/documents/"+d.ID+"/annotations", []byte(`{"1":[{"kind":"redact","x":-1}]}`), nil).Code != 400 {
		t.Fatal("invalid geometry")
	}
	if m.call("PUT", "/api/documents/"+d.ID+"/annotations", []byte(`{}`), nil).Code != 404 {
		t.Fatal("annotation ACL")
	}
	if o.call("PUT", "/api/documents/"+d.ID+"/annotations", []byte(`{"1":[{"kind":"redact","x":0.1,"y":0.4,"w":0.8,"h":0.2}]}`), nil).Code != 200 {
		t.Fatal("annotation persistence")
	}
	in, _ := json.Marshal(Request{Operation: "rotate", Documents: []string{d.ID}, Rotation: 90})
	w = o.call("POST", "/api/jobs", in, nil)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body)
	}
	var job map[string]string
	json.Unmarshal(w.Body.Bytes(), &job)
	if m.call("DELETE", "/api/jobs/"+job["id"], nil, nil).Code != 404 {
		t.Fatal("job ACL")
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	deadline := time.Now().Add(20 * time.Second)
	state := ""
	for time.Now().Before(deadline) {
		h.DB.QueryRow("SELECT state FROM pdf_jobs WHERE id=?", job["id"]).Scan(&state)
		if state == "complete" || state == "failed" {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}
	cancel()
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if state != "complete" {
		t.Fatal("job failed", state)
	}
	if o.call("GET", "/api/documents/"+d.ID, nil, nil).Code != 200 {
		t.Fatal("source removed")
	}
}
