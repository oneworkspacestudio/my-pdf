package main

import (
	"github.com/oneworkspacestudio/my-pdf/internal/app"
	"github.com/oneworkspacestudio/my-pdf/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
)

var version = "0.1.0-preview.2"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--worker" {
		if e := app.Worker(os.Stdin, os.Stdout); e != nil {
			os.Exit(2)
		}
		return
	}
	if e := apphost.Main(apphost.Config{ID: "pdf", Name: "My PDF", Version: version, Listen: "127.0.0.1:9009"}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return app.New(h) }); e != nil {
		apphost.Report(e)
		os.Exit(1)
	}
}
