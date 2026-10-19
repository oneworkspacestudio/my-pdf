package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Mark struct {
	Kind   string      `json:"kind"`
	X      float64     `json:"x"`
	Y      float64     `json:"y"`
	W      float64     `json:"w"`
	H      float64     `json:"h"`
	Text   string      `json:"text"`
	Image  string      `json:"image"`
	Points [][]float64 `json:"points"`
}

func (a *App) annotations(w http.ResponseWriter, r *http.Request) {
	if _, e := a.doc(apphost.Current(r).User.ID, r.PathValue("id")); e != nil {
		apphost.Problem(w, 404, "Document not found")
		return
	}
	if r.Method == "GET" {
		var v string
		if a.H.DB.QueryRow("SELECT value FROM pdf_annotations WHERE document_id=?", r.PathValue("id")).Scan(&v) != nil {
			v = "{}"
		}
		apphost.JSON(w, 200, json.RawMessage(v))
		return
	}
	var marks map[string][]Mark
	if !apphost.Decode(w, r, &marks) {
		return
	}
	total := 0
	for key, page := range marks {
		n, e := strconv.Atoi(key)
		if e != nil || n < 1 || n > 1000 {
			apphost.Problem(w, 400, "Invalid page number")
			return
		}
		total += len(page)
		if total > 1000 {
			apphost.Problem(w, 413, "Keep up to 1000 marks per document")
			return
		}
		for _, m := range page {
			switch m.Kind {
			case "text", "comment", "stamp", "image", "highlight", "shape", "ink", "redact":
			default:
				apphost.Problem(w, 400, "Unsupported annotation")
				return
			}
			if len(m.Text) > 2000 || len(m.Points) > 2000 {
				apphost.Problem(w, 413, "Annotation is too large")
				return
			}
			if m.Image != "" {
				if !strings.HasPrefix(m.Image, "data:image/png;base64,") || len(m.Image) > 180000 {
					apphost.Problem(w, 400, "Stamp must be a small PNG image")
					return
				}
				raw, e := base64.StdEncoding.DecodeString(strings.TrimPrefix(m.Image, "data:image/png;base64,"))
				if e != nil {
					apphost.Problem(w, 400, "Invalid stamp image")
					return
				}
				cfg, e := png.DecodeConfig(bytes.NewReader(raw))
				if e != nil || cfg.Width > 1024 || cfg.Height > 1024 {
					apphost.Problem(w, 400, "Stamp image exceeds 1024 pixels")
					return
				}
			}
			coords := []float64{m.X, m.Y, m.W, m.H}
			for _, point := range m.Points {
				if len(point) != 2 {
					apphost.Problem(w, 400, "Invalid ink point")
					return
				}
				coords = append(coords, point...)
			}
			for _, v := range coords {
				if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
					apphost.Problem(w, 400, "Annotation coordinates must be on the page")
					return
				}
			}
		}
	}
	v, _ := json.Marshal(marks)
	_, e := a.H.DB.Exec("INSERT INTO pdf_annotations VALUES(?,?) ON CONFLICT(document_id)DO UPDATE SET value=excluded.value", r.PathValue("id"), string(v))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
