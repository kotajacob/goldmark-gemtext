package gemtext

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"

	wiki "git.sr.ht/~kota/goldmark-wiki"
	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func setupFiles(srcPath, wantPath string) (src, want []byte, err error) {
	src, err = os.ReadFile(srcPath)
	if err != nil {
		return src, nil, err
	}
	want, err = os.ReadFile(wantPath)
	return src, want, err
}

// TestNew runs a test for each configutation option by creating a GemRenderer
// with New and applying options using the WithOption() functions.
func TestNew(t *testing.T) {
	tests := []struct {
		srcPath  string
		wantPath string
		option   Option
	}{
		{
			"test_data/render.md", "test_data/renderDefault.gmi",
			WithHeadingLink(HeadingLinkAuto),
		},
		{
			"test_data/render.md", "test_data/renderHeadingSpaceSingle.gmi",
			WithHeadingSpace(HeadingSpaceSingle),
		},
		{
			"test_data/render.md", "test_data/renderParagraphLinkOff.gmi",
			WithParagraphLink(ParagraphLinkOff),
		},
		{
			"test_data/render.md", "test_data/renderParagraphLinkCurlyBelow.gmi",
			WithParagraphLink(ParagraphLinkCurlyBelow),
		},

		{
			"test_data/render.md", "test_data/renderHeadLinkOff.gmi",
			WithHeadingLink(HeadingLinkOff),
		},
		{
			"test_data/render.md", "test_data/renderHeadLinkBelow.gmi",
			WithHeadingLink(HeadingLinkBelow),
		},
		{
			"test_data/render.md", "test_data/renderEmphasisMarkdown.gmi",
			WithEmphasis(EmphasisMarkdown),
		},
		{
			"test_data/render.md", "test_data/renderEmphasisUnicode.gmi",
			WithEmphasis(EmphasisUnicode),
		},
		{
			"test_data/render.md", "test_data/renderCodeSpanMarkdown.gmi",
			WithCodeSpan(CodeSpanMarkdown),
		},
		{
			"test_data/render.md", "test_data/renderStrikethroughMarkdown.gmi",
			WithStrikethrough(StrikethroughMarkdown),
		},
		{
			"test_data/render.md", "test_data/renderStrikethroughUnicode.gmi",
			WithStrikethrough(StrikethroughUnicode),
		},
		{
			"test_data/render.md", "test_data/renderHorizontalRule.gmi",
			WithHorizontalRule("+++"),
		},
		{
			"test_data/render.md", "test_data/renderLinkReplacers.gmi",
			WithLinkReplacers([]LinkReplacer{
				{
					LinkMarkdown,
					regexp.MustCompile(`https?`),
					"markdownlink",
				},
				{
					LinkWiki,
					regexp.MustCompile(`nz`),
					"org",
				},
				{
					LinkAuto,
					regexp.MustCompile(`https?`),
					"autolinks",
				},
				{
					LinkImage,
					regexp.MustCompile(`https?`),
					"image",
				},
			}),
		},
	}

	for _, test := range tests {
		src, want, err := setupFiles(test.srcPath, test.wantPath)
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithExtensions(
				wiki.Wiki,
				extension.Linkify,
				extension.Strikethrough,
			),
		)

		md.SetRenderer(New(test.option))
		if err := md.Convert(src, &buf); err != nil {
			t.Fatal(err)
		}
		got := buf.Bytes()

		if !cmp.Equal(got, want) {
			err := os.WriteFile("fail.gmi", got, 0644)
			if err != nil {
				t.Fatal(err)
			}
			t.Fatal(cmp.Diff(got, want))
		}
	}
}

// TestNewGemRenderer creates a GemRenderer by manually creating a Config and
// passing it into NewGemRenderer. Only the default configuration is tested as
// all configuration options are tested by TestNew.
func TestNewGemRenderer(t *testing.T) {
	tests := []struct {
		srcPath  string
		wantPath string
		config   Config
	}{
		{
			"test_data/render.md", "test_data/renderDefault.gmi",
			Config{HeadingLinkAuto, HeadingSpaceDouble, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanOff, HR, []LinkReplacer{}},
		},
	}

	for _, test := range tests {
		src, want, err := setupFiles(test.srcPath, test.wantPath)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithExtensions(
				wiki.Wiki,
				extension.Linkify,
				extension.Strikethrough,
			),
		)

		ar := NewGemRenderer(&test.config)
		md.SetRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

		if err := md.Convert(src, &buf); err != nil {
			t.Fatal(err)
		}
		got := buf.Bytes()

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
