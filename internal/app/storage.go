package app

import (
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const userQuota int64 = 1 << 30

func (a *App) budget(user string) int64 {
	var used int64
	a.H.DB.QueryRow("SELECT (SELECT COALESCE(SUM(size),0) FROM pdf_documents WHERE user_id=?)+(SELECT COALESCE(SUM(size),0) FROM pdf_artifacts WHERE user_id=?)", user, user).Scan(&used)
	return max(int64(0), userQuota-used)
}
func (a *App) gcFiles() {
	rows, e := a.H.DB.Query("SELECT hash FROM pdf_documents UNION SELECT hash FROM pdf_artifacts")
	if e != nil {
		return
	}
	live := map[string]bool{}
	for rows.Next() {
		var hash string
		rows.Scan(&hash)
		live[hash] = true
	}
	rows.Close()
	entries, e := os.ReadDir(filepath.Join(a.H.Config.Data, "objects"))
	if e != nil {
		return
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || live[entry.Name()] {
			continue
		}
		if path, e := a.H.BlobPath(entry.Name()); e == nil {
			os.Remove(path)
		}
	}
}
func (a *App) cleanJobs() {
	entries, e := os.ReadDir(a.H.Config.Data)
	if e != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "pdf-work-") {
			os.RemoveAll(filepath.Join(a.H.Config.Data, entry.Name()))
		}
	}
}
func (a *App) deleteArtifact(w http.ResponseWriter, r *http.Request) {
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	u := apphost.Current(r).User.ID
	_, e := a.H.DB.Exec("DELETE FROM pdf_artifacts WHERE id=? AND user_id=?", r.PathValue("id"), u)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.DB.Exec("UPDATE pdf_jobs SET result='',message='Result removed; originals are unchanged' WHERE result=? AND user_id=?", r.PathValue("id"), u)
	a.gcFiles()
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
