package gemtext

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func setupFiles(srcPath, wantPath string) (err error, src, want []byte) {
	src, err = os.ReadFile(srcPath)
	if err != nil {
		return err, nil, nil
	}
	want, err = os.ReadFile(wantPath)
	if err != nil {
		return err, nil, nil
	}
	return nil, src, want
}

func runNew(srcPath string, wantPath string, option Option) (error, []byte, []byte) {
	err, src, want := setupFiles(srcPath, wantPath)
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
	md.SetRenderer(New(option))
	if err := md.Convert(src, &buf); err != nil {
		return err, nil, nil
	}
	got := buf.Bytes()
	return nil, want, got
}

func runNewGemRenderer(srcPath string, wantPath string, config Config) (error, []byte, []byte) {
	err, src, want := setupFiles(srcPath, wantPath)
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

func TestNew(t *testing.T) {
	var tests = []struct {
		srcPath  string
		wantPath string
		option   Option
	}{
		{"test_data/render.md", "test_data/renderDefault.gmi",
			WithHeadingLink(HeadingLinkAuto)},
		{"test_data/render.md", "test_data/renderHeadLinkOff.gmi",
			WithHeadingLink(HeadingLinkOff)},
		{"test_data/render.md", "test_data/renderEmphasisMarkdown.gmi",
			WithEmphasis(EmphasisMarkdown)},
		{"test_data/render.md", "test_data/renderEmphasisUnicode.gmi",
			WithEmphasis(EmphasisUnicode)},
		{"test_data/render.md", "test_data/renderCodeSpanMarkdown.gmi",
			WithCodeSpan(CodeSpanMarkdown)},
		{"test_data/render.md", "test_data/renderStrikethroughMarkdown.gmi",
			WithStrikethrough(StrikethroughMarkdown)},
		{"test_data/render.md", "test_data/renderStrikethroughUnicode.gmi",
			WithStrikethrough(StrikethroughUnicode)},
	}

	for _, test := range tests {
		err, want, got := runNew(test.srcPath, test.wantPath, test.option)
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

func TestNewGemRenderer(t *testing.T) {
	var tests = []struct {
		srcPath  string
		wantPath string
		config   Config
	}{
		{"test_data/render.md", "test_data/renderDefault.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanOff}},
		{"test_data/render.md", "test_data/renderHeadLinkOff.gmi",
			Config{HeadingLinkOff, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanOff}},
		{"test_data/render.md", "test_data/renderEmphasisMarkdown.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisMarkdown, StrikethroughOff, CodeSpanOff}},
		{"test_data/render.md", "test_data/renderEmphasisUnicode.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisUnicode, StrikethroughOff, CodeSpanOff}},
		{"test_data/render.md", "test_data/renderCodeSpanMarkdown.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanMarkdown}},
		{"test_data/render.md", "test_data/renderStrikethroughMarkdown.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisOff, StrikethroughMarkdown, CodeSpanOff}},
		{"test_data/render.md", "test_data/renderStrikethroughUnicode.gmi",
			Config{HeadingLinkAuto, ParagraphLinkBelow, EmphasisOff, StrikethroughUnicode, CodeSpanOff}},
	}

	for _, test := range tests {
		err, want, got := runNewGemRenderer(test.srcPath, test.wantPath, test.config)
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

func BenchmarkConvert(b *testing.B) {
	srcPath := "test_data/render.md"
	src, err := os.ReadFile(srcPath)
	if err != nil {
		b.Fatalf("failed to load testing data: %v", err)
	}
	buf := new(bytes.Buffer)
	w := iotest.TruncateWriter(buf, 0) // no need to actually store the data
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	md.SetRenderer(New())
	for i := 0; i < b.N; i++ {
		if err := md.Convert(src, w); err != nil {
			b.Fatalf("failed running benchmark: %v", err)
		}
	}
}
