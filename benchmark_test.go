package gemtext

import (
	"bytes"
	"os"
	"testing"
	"testing/iotest"

	wiki "git.sr.ht/~kota/goldmark-wiki"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

func benchCreateRenderer() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(
			wiki.Wiki,
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	md.SetRenderer(New())
	return md
}

func BenchmarkConvert(b *testing.B) {
	srcPath := "test_data/render.md"
	src, err := os.ReadFile(srcPath)
	if err != nil {
		b.Fatalf("failed to load testing data: %v", err)
	}
	buf := new(bytes.Buffer)
	w := iotest.TruncateWriter(buf, 0) // no need to actually store the data
	md := benchCreateRenderer()
	for i := 0; i < b.N; i++ {
		if err := md.Convert(src, w); err != nil {
			b.Fatalf("failed running benchmark: %v", err)
		}
	}
}

func BenchmarkRender(b *testing.B) {
	srcPath := "test_data/render.md"
	src, err := os.ReadFile(srcPath)
	if err != nil {
		b.Fatalf("failed to load testing data: %v", err)
	}
	buf := new(bytes.Buffer)
	w := iotest.TruncateWriter(buf, 0) // no need to actually store the data
	md := benchCreateRenderer()
	reader := text.NewReader(src)
	par := md.Parser()
	doc := par.Parse(reader)
	ren := md.Renderer()
	for i := 0; i < b.N; i++ {
		ren.Render(w, src, doc)
	}
}
