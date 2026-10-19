package app

import (
	"context"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"os"
)

func documentInfo(c context.Context, path string, conf *model.Configuration) (map[string]any, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	ctx, e := api.ReadValidateAndOptimize(c, f, conf, nil)
	if e != nil {
		return nil, e
	}
	return map[string]any{"title": ctx.Title, "author": ctx.Author, "subject": ctx.Subject, "keywords": ctx.Keywords, "creator": ctx.Creator, "producer": ctx.Producer, "created": ctx.XRefTable.CreationDate, "modified": ctx.ModDate, "pages": ctx.PageCount, "properties": ctx.Properties, "signaturesPresent": ctx.SignatureExist}, nil
}

// Remove Info/XMP objects themselves, not just their dictionary links. Other
// document content (attachments, annotations, forms) is not claimed sanitised.
func removeDocumentMetadata(c context.Context, in, out string, conf *model.Configuration) error {
	f, e := os.Open(in)
	if e != nil {
		return e
	}
	defer f.Close()
	ctx, e := api.ReadValidateAndOptimize(c, f, conf, nil)
	if e != nil {
		return e
	}
	if ctx.Info != nil {
		if e = ctx.DeleteObjectGraph(c, *ctx.Info); e != nil {
			return e
		}
		ctx.Info = nil
	}
	ctx.DirectInfoDict = nil
	ctx.Title = ""
	ctx.Author = ""
	ctx.Subject = ""
	ctx.Keywords = ""
	ctx.Creator = ""
	ctx.Producer = ""
	ctx.XRefTable.CreationDate = ""
	ctx.ModDate = ""
	ctx.Properties = map[string]string{}
	ctx.CatalogXMPMeta = nil
	catalog, e := ctx.Catalog()
	if e != nil {
		return e
	}
	remove := func(d types.Dict) error {
		if object, ok := d["Metadata"]; ok {
			delete(d, "Metadata")
			return ctx.DeleteObjectGraph(c, object)
		}
		return nil
	}
	if e = remove(catalog); e != nil {
		return e
	}
	for n := 1; n <= ctx.PageCount; n++ {
		page, _, _, e := ctx.PageDict(c, n, false)
		if e != nil {
			return e
		}
		if e = remove(page); e != nil {
			return e
		}
	}
	output, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer output.Close()
	return api.Write(c, ctx, output, conf)
}
