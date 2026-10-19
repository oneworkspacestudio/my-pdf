package app

import (
	"archive/zip"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func zipPages(dir, out string) error {
	files, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	if len(files) == 0 || len(files) > 1000 {
		return errors.New("split page count exceeds limit")
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	z := zip.NewWriter(f)
	var size int64
	for _, entry := range files {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".pdf") || strings.ContainsAny(entry.Name(), "/\\") {
			z.Close()
			return errors.New("unexpected split output")
		}
		path := filepath.Join(dir, entry.Name())
		info, e := entry.Info()
		if e != nil {
			z.Close()
			return e
		}
		size += info.Size()
		if size > maxFile {
			z.Close()
			return errors.New("split archive exceeds limit")
		}
		source, e := os.Open(path)
		if e != nil {
			z.Close()
			return e
		}
		dest, e := z.Create(entry.Name())
		if e == nil {
			_, e = io.Copy(dest, source)
		}
		source.Close()
		if e != nil {
			z.Close()
			return e
		}
	}
	return z.Close()
}
func (a *App) artifact(w http.ResponseWriter, r *http.Request) {
	var hash string
	var size int64
	if a.H.DB.QueryRow("SELECT hash,size FROM pdf_artifacts WHERE id=? AND user_id=?", r.PathValue("id"), apphost.Current(r).User.ID).Scan(&hash, &size) != nil {
		apphost.Problem(w, 404, "Artifact not found")
		return
	}
	path, e := a.H.BlobPath(hash)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="split-pages.zip"`)
	http.ServeFile(w, r, path)
}
