package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"os"
	"path/filepath"
	"testing"
)

func runWorker(t *testing.T, dir string, in WorkerRequest) (map[string]any, error) {
	t.Helper()
	raw, _ := json.Marshal(in)
	exe, _ := os.Executable()
	var out bytes.Buffer
	e := apphost.Worker(context.Background(), exe, dir, bytes.NewReader(raw), &out)
	var result map[string]any
	json.Unmarshal(out.Bytes(), &result)
	return result, e
}
func TestEncryptedSignedAndMetadataCopies(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.pdf")
	os.WriteFile(source, fixturePDF(1), 0600)
	enc := filepath.Join(dir, "encrypted.pdf")
	conf := model.NewStatelessConfiguration()
	conf.Offline = true
	conf.UserPW = "fixture-user-password"
	conf.OwnerPW = "fixture-owner-password"
	conf.EncryptUsingAES = true
	conf.EncryptKeyLength = 256
	if e := api.EncryptFile(context.Background(), source, enc, conf); e != nil {
		t.Fatal(e)
	}
	wr := WorkerRequest{Operation: "inspect", Inputs: []string{enc}, Output: filepath.Join(dir, "result.pdf")}
	if _, e := runWorker(t, dir, wr); e == nil {
		t.Fatal("encrypted PDF opened without password")
	}
	wr.Password = "fixture-user-password"
	if _, e := runWorker(t, dir, wr); e != nil {
		t.Fatal("correct PDF password rejected", e)
	}
	wr.Password = "wrong-password"
	if _, e := runWorker(t, dir, wr); e == nil {
		t.Fatal("wrong password accepted")
	}
	signed, e := os.ReadFile("testdata/signed-selfsigned.pdf")
	if e != nil {
		t.Fatal(e)
	}
	signedPath := filepath.Join(dir, "signed.pdf")
	os.WriteFile(signedPath, signed, 0600)
	wr = WorkerRequest{Operation: "inspect", Inputs: []string{signedPath}, Output: filepath.Join(dir, "signed-copy.pdf")}
	result, e := runWorker(t, dir, wr)
	if e != nil {
		t.Fatal(e)
	}
	document, _ := result["document"].(map[string]any)
	if document["signaturesPresent"] != true {
		t.Fatal("signature not detected", result)
	}
	after, _ := os.ReadFile(signedPath)
	if !bytes.Equal(signed, after) {
		t.Fatal("signed original changed")
	}
	wr = WorkerRequest{Operation: "metadata-remove", Inputs: []string{source}, Output: filepath.Join(dir, "no-metadata.pdf")}
	if _, e = runWorker(t, dir, wr); e != nil {
		t.Fatal("metadata operation", e)
	}
	info, e := documentInfo(context.Background(), wr.Output, model.NewStatelessConfiguration())
	if e != nil || info["author"] != "" || info["title"] != "" {
		t.Fatal(info, e)
	}
}
