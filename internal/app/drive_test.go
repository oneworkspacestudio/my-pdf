package app

import (
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"path/filepath"
	"testing"
)

func TestDriveQueueLeasePersistenceAndOwnership(t *testing.T) {
	h, e := apphost.New(apphost.Config{ID: "pdf", Name: "My PDF", Data: filepath.Join(t.TempDir(), "data"), Origin: "http://localhost:29009", Listen: "127.0.0.1:29009"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	if _, e = New(h); e != nil {
		t.Fatal(e)
	}
	h.Bootstrap("owner@example.test", "fixture-password")
	o := &pdfClient{h: h}
	pdfLogin(t, o, "owner@example.test")
	o.call("POST", "/api/local-users", []byte(`{"email":"member@example.test","password":"fixture-password"}`), nil)
	m := &pdfClient{h: h}
	pdfLogin(t, m, "member@example.test")
	w := o.call("POST", "/api/documents", fixturePDF(1), map[string]string{"X-File-Name": "queued.pdf", "X-Save-To-Drive": "true"})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body)
	}
	var d Document
	json.Unmarshal(w.Body.Bytes(), &d)
	endpoint := "/api/documents/" + d.ID + "/drive-save"
	call := func(body string, want int) {
		t.Helper()
		w := o.call("POST", endpoint, []byte(body), nil)
		if w.Code != want {
			t.Fatal(w.Code, w.Body)
		}
	}
	if m.call("POST", endpoint, []byte(`{"action":"claim"}`), nil).Code != 404 {
		t.Fatal("cross-user claim")
	}
	w = o.call("POST", endpoint, []byte(`{"action":"claim"}`), nil)
	var lease map[string]string
	json.Unmarshal(w.Body.Bytes(), &lease)
	if w.Code != 200 || lease["lease"] == "" {
		t.Fatal(w.Code, w.Body)
	}
	call(`{"action":"claim"}`, 409)
	call(`{"action":"saved","lease":"wrong"}`, 409)
	target := `{"action":"target","lease":"` + lease["lease"] + `","source":"My Drive","path":"/My PDF/queued.pdf"}`
	call(target, 200)
	call(`{"action":"target","lease":"`+lease["lease"]+`","source":"My Drive","path":"/../escape.pdf"}`, 400)
	h.DB.Exec("UPDATE pdf_drive SET expires=0 WHERE document_id=?", d.ID)
	w = o.call("POST", endpoint, []byte(`{"action":"claim"}`), nil)
	json.Unmarshal(w.Body.Bytes(), &lease)
	if w.Code != 200 || lease["path"] != "/My PDF/queued.pdf" {
		t.Fatal("expired claim lost destination", w.Body)
	}
	call(`{"action":"saved","lease":"`+lease["lease"]+`"}`, 200)
	call(`{"action":"claim"}`, 409)
	var docs []Document
	json.Unmarshal(o.call("GET", "/api/documents", nil, nil).Body.Bytes(), &docs)
	if len(docs) != 1 || docs[0].Drive.State != "saved" || docs[0].Drive.Path != "/My PDF/queued.pdf" {
		t.Fatal(docs)
	}
	h.Close()
	h, e = apphost.New(h.Config, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	if _, e = New(h); e != nil {
		t.Fatal("schema restart", e)
	}
	o.h = h
	pdfLogin(t, o, "owner@example.test")
	json.Unmarshal(o.call("GET", "/api/documents", nil, nil).Body.Bytes(), &docs)
	if len(docs) != 1 || docs[0].Drive.State != "saved" {
		t.Fatal("receipt lost on restart", docs)
	}
	o.call("DELETE", "/api/documents/"+d.ID, nil, nil)
	var n int
	h.DB.QueryRow("SELECT COUNT(*) FROM pdf_drive").Scan(&n)
	if n != 0 {
		t.Fatal("orphan receipt")
	}
}
