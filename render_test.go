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

func run(srcPath string, wantPath string, config Config) (error, []byte, []byte) {
	src, err := os.ReadFile(srcPath)
	if err != nil {
		return err, nil, nil
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		return err, nil, nil
	}
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	ar := NewGemRenderer(&config)
	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

	if err := md.Convert(src, &buf); err != nil {
		return err, nil, nil
	}
	got := buf.Bytes()
	return nil, want, got
}

func TestRender(t *testing.T) {
	var tests = []struct {
		srcPath  string
		wantPath string
		config   Config
	}{
		{"test_data/render.md", "test_data/renderDefault.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanOff}},
	}

	for _, test := range tests {
		err, want, got := run(test.srcPath, test.wantPath, test.config)
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			err := os.WriteFile("fail.gmi", got, 0644)
			if err != nil {
				t.Fatal(err)
			}
			t.Fatal(fmt.Println(cmp.Diff(got, want)))
		}
	}
}
