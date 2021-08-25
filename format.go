package gemtext

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

func Format(source []byte, w io.Writer, opts ...parser.ParseOption) error {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	doc := md.Parser().Parse(
		text.NewReader(source), opts...)
	return Render(w, source, doc)
}

// Gemtext is a gemtext format renderer.
var Gemtext renderer.Renderer = new(gemtextRenderer)

type gemtextRenderer struct{}

// AddOptions adds given option to this renderer.
func (*gemtextRenderer) AddOptions(opts ...renderer.Option) {}

// Write render node as Gemtext.
func (*gemtextRenderer) Render(w io.Writer, source []byte, node ast.Node) (err error) {
	return Render(w, source, node)
}
