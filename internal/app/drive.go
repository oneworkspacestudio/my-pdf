package app

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

// Receipts describe ordinary Drive copies, never grant access to Drive. All file
// operations still run through the browser's independently checked Drive user.
type DriveSave struct {
	State   string `json:"state"`
	Source  string `json:"source"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func validDriveTarget(source, target string) bool {
	return source != "" && len(source) <= 200 && len(target) <= 2000 &&
		strings.HasPrefix(target, "/") && target != "/" && path.Clean(target) == target &&
		!strings.ContainsAny(source+target, "\\\x00\r\n") && strings.HasSuffix(strings.ToLower(target), ".pdf")
}

func (a *App) driveSave(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, e := a.doc(apphost.Current(r).User.ID, id); e != nil {
		apphost.Problem(w, 404, "Document not found")
		return
	}
	var in struct{ Action, Lease, Source, Path string }
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); e != nil {
		apphost.Problem(w, 400, "Invalid Drive save request")
		return
	}
	now := time.Now().Unix()
	var result any = map[string]bool{"ok": true}
	var query string
	var args []any
	switch in.Action {
	case "blocked":
		query = `UPDATE pdf_drive SET state='error',message='Drive connection required. Retry saving to connect your account.' WHERE document_id=? AND state='pending'`
		args = []any{id}
	case "retry":
		query = `INSERT INTO pdf_drive(document_id,state) VALUES(?,'pending') ON CONFLICT(document_id) DO UPDATE SET state='pending',message='',lease='',expires=0 WHERE state='error'`
		args = []any{id}
	case "claim":
		token := apphost.ID()
		query = `UPDATE pdf_drive SET state='saving',lease=?,expires=?,message='' WHERE document_id=? AND (state='pending' OR (state='saving' AND expires<?))`
		args = []any{token, now + 300, id, now}
		result = map[string]string{"lease": token}
	case "claimCopy":
		token := apphost.ID()
		query = `INSERT INTO pdf_drive(document_id,state,lease,expires) VALUES(?,'saving',?,?) ON CONFLICT(document_id) DO UPDATE SET state='saving',lease=excluded.lease,expires=excluded.expires,message='',source='',path='' WHERE state<>'saving' OR expires<?`
		args = []any{id, token, now + 300, now}
		result = map[string]string{"lease": token}
	case "target":
		if !validDriveTarget(in.Source, in.Path) {
			apphost.Problem(w, 400, "Choose a valid Drive PDF path")
			return
		}
		query = `UPDATE pdf_drive SET source=?,path=? WHERE document_id=? AND state='saving' AND lease=? AND expires>=?`
		args = []any{in.Source, in.Path, id, in.Lease, now}
	case "saved":
		query = `UPDATE pdf_drive SET state='saved',message='',lease='',expires=0 WHERE document_id=? AND state='saving' AND lease=? AND expires>=? AND source<>'' AND path<>''`
		args = []any{id, in.Lease, now}
	case "failed":
		query = `UPDATE pdf_drive SET state='error',message='Not saved to Drive. Check the connection and folder permissions, then retry.',lease='',expires=0 WHERE document_id=? AND state='saving' AND lease=?`
		args = []any{id, in.Lease}
	case "defer":
		query = `UPDATE pdf_drive SET state='pending',lease='',expires=0 WHERE document_id=? AND state='saving' AND lease=?`
		args = []any{id, in.Lease}
	default:
		apphost.Problem(w, 400, "Unsupported Drive save action")
		return
	}
	res, e := a.H.DB.Exec(query, args...)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		apphost.Problem(w, 409, "This document is already saved or another tab is saving it. Refresh its status.")
		return
	}
	if in.Action == "claim" || in.Action == "claimCopy" {
		var source, target string
		a.H.DB.QueryRow("SELECT source,path FROM pdf_drive WHERE document_id=?", id).Scan(&source, &target)
		result.(map[string]string)["source"] = source
		result.(map[string]string)["path"] = target
	}
	apphost.JSON(w, 200, result)
}
