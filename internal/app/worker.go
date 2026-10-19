package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"io"
	"os"
	"path/filepath"
)

type WorkerRequest struct {
	Operation string
	Inputs    []string
	Output    string
	Pages     []string
	Rotation  int
	Password  string
}

// The worker accepts only files in its dedicated working directory. The parent
// supplies paths after checking document ownership; no browser path reaches here.
func Worker(stdin io.Reader, stdout io.Writer) error {
	var in WorkerRequest
	dec := json.NewDecoder(io.LimitReader(stdin, 1<<20))
	dec.DisallowUnknownFields()
	if e := dec.Decode(&in); e != nil {
		return e
	}
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	for _, p := range append(append([]string{}, in.Inputs...), in.Output) {
		abs, e := filepath.Abs(p)
		if e != nil || filepath.Dir(abs) != cwd {
			return errors.New("worker path outside job folder")
		}
	}
	request := Request{Operation: in.Operation, Documents: in.Inputs, Pages: in.Pages, Rotation: in.Rotation, Password: in.Password}
	if e = validRequest(request); e != nil {
		return e
	}
	api.DisableConfigDir()
	conf := model.NewStatelessConfiguration()
	conf.UserPW = in.Password
	conf.OwnerPW = in.Password
	conf.Offline = true
	conf.CheckFileNameExt = false
	ctx := context.Background()
	metadata := map[string]any{"engine": "pdfcpu", "originalsUnchanged": true}
	total := 0
	for _, path := range in.Inputs {
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		n, e := api.PageCount(ctx, f, conf)
		f.Close()
		if e != nil {
			return e
		}
		total += n
		if total > 1000 {
			return errors.New("maximum 1000 pages")
		}
	}
	metadata["pages"] = total
	switch in.Operation {
	case "inspect":
		var info map[string]any
		info, e = documentInfo(ctx, in.Inputs[0], conf)
		metadata["document"] = info
	case "split":
		dir := filepath.Join(cwd, "split")
		if e = os.Mkdir(dir, 0700); e == nil {
			e = api.SplitFile(ctx, in.Inputs[0], dir, 1, conf)
		}
	case "merge":
		e = api.MergeCreateFile(ctx, in.Inputs, in.Output, false, conf)
	case "extract":
		e = api.CollectFile(ctx, in.Inputs[0], in.Output, in.Pages, conf)
	case "remove":
		e = api.RemovePagesFile(ctx, in.Inputs[0], in.Output, in.Pages, conf)
	case "rotate":
		e = api.RotateFile(ctx, in.Inputs[0], in.Output, in.Rotation, in.Pages, conf)
	case "compress":
		e = api.OptimizeFile(ctx, in.Inputs[0], in.Output, conf, nil)
	case "blank":
		e = api.InsertPagesFile(ctx, in.Inputs[0], in.Output, in.Pages, false, nil, conf)
	case "metadata-remove":
		e = removeDocumentMetadata(ctx, in.Inputs[0], in.Output, conf)
	}
	if e != nil {
		return e
	}
	return json.NewEncoder(stdout).Encode(metadata)
}

type limitedBuffer struct{ bytes.Buffer }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 1<<20 {
		return 0, errors.New("worker response limit")
	}
	return b.Buffer.Write(p)
}
