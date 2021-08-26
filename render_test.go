package gemtext

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func TestRender(t *testing.T) {
	src, err := os.ReadFile("test_data/render.md")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_data/render.gmi")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer

	config := NewConfig()
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	ar := NewGemRenderer(config)
	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

	if err := md.Convert(src, &buf); err != nil {
		t.Fatal(err)
	}

	got := buf.Bytes()
	if !cmp.Equal(got, want) {
		err := os.WriteFile("fail.gmi", got, 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(fmt.Println(cmp.Diff(got, want)))
	}
}
