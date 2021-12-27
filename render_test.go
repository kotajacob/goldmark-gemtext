package gemtext

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"testing/iotest"

	wiki "git.sr.ht/~kota/goldmark-wiki"
	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func ExampleNew() {
	src := `
# This is a heading

This is a [paragraph](https://en.wikipedia.org/wiki/Paragraph) with [some
links](https://en.wikipedia.org/wiki/Hyperlink) in it.

Next we'll have a list of some musicians I like, but as an individual list of
links. One of the neat features of goldmark-gemtext is that it recognizes when
a "paragraph" is really just a list of links and handles it as if it's a list
of links by simply converting them to the gemtext format. I wasn't able to find
any other markdown to gemtext tools that could do this so it was the
inspiration for writing this in the first place.

[Noname](https://nonameraps.bandcamp.com/)\
[Milo](https://afrolab9000.bandcamp.com/album/so-the-flies-dont-come)\
[Busdriver](https://busdriver-thumbs.bandcamp.com/)\
[Neat Beats](https://www.youtube.com/watch?v=X6kGg31G0As)\
[Ratatat](http://www.ratatatmusic.com/)\
[Sylvan Esso](https://www.sylvanesso.com/)\
[Phoebe Bridgers](https://phoebefuckingbridgers.com/)
`
	// create markdown parser
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)

	// set some options
	options := []Option{WithHeadingLink(HeadingLinkAuto), WithCodeSpan(CodeSpanMarkdown)}

	md.SetRenderer(New(options...))
	_ = md.Convert([]byte(src), &buf) // ignoring errors for example
	fmt.Println(buf.String())
	// Output:
	// # This is a heading
	//
	// This is a paragraph with some links in it.
	//
	// => https://en.wikipedia.org/wiki/Paragraph paragraph
	// => https://en.wikipedia.org/wiki/Hyperlink some links
	//
	// Next we'll have a list of some musicians I like, but as an individual list of links. One of the neat features of goldmark-gemtext is that it recognizes when a "paragraph" is really just a list of links and handles it as if it's a list of links by simply converting them to the gemtext format. I wasn't able to find any other markdown to gemtext tools that could do this so it was the inspiration for writing this in the first place.
	//
	// => https://nonameraps.bandcamp.com/ Noname
	// => https://afrolab9000.bandcamp.com/album/so-the-flies-dont-come Milo
	// => https://busdriver-thumbs.bandcamp.com/ Busdriver
	// => https://www.youtube.com/watch?v=X6kGg31G0As Neat Beats
	// => http://www.ratatatmusic.com/ Ratatat
	// => https://www.sylvanesso.com/ Sylvan Esso
	// => https://phoebefuckingbridgers.com/ Phoebe Bridgers
}

func setupFiles(srcPath, wantPath string) (src, want []byte, err error) {
	src, err = os.ReadFile(srcPath)
	if err != nil {
		return src, nil, err
	}
	want, err = os.ReadFile(wantPath)
	return src, want, err
}

func runNew(srcPath string, wantPath string, option Option) ([]byte, []byte, error) {
	src, want, err := setupFiles(srcPath, wantPath)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			wiki.Wiki,
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	md.SetRenderer(New(option))
	if err := md.Convert(src, &buf); err != nil {
		return nil, nil, err
	}
	got := buf.Bytes()
	return want, got, err
}

func runNewGemRenderer(srcPath string, wantPath string, config Config) ([]byte, []byte, error) {
	src, want, err := setupFiles(srcPath, wantPath)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			wiki.Wiki,
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	ar := NewGemRenderer(&config)
	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

	if err := md.Convert(src, &buf); err != nil {
		return nil, nil, err
	}
	got := buf.Bytes()
	return want, got, err
}

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
	}

	for _, test := range tests {
		want, got, err := runNew(test.srcPath, test.wantPath, test.option)
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
	tests := []struct {
		srcPath  string
		wantPath string
		config   Config
	}{
		{
			"test_data/render.md", "test_data/renderDefault.gmi",
			Config{HeadingLinkAuto, HeadingSpaceDouble, ParagraphLinkBelow, EmphasisOff, StrikethroughOff, CodeSpanOff, HR},
		},
	}

	for _, test := range tests {
		want, got, err := runNewGemRenderer(test.srcPath, test.wantPath, test.config)
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
