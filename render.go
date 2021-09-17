package gemtext

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// New returns a gemtext renderer.
func New(opts ...Option) renderer.Renderer {
	config := NewConfig()
	for _, opt := range opts {
		opt.SetConfig(config)
	}
	r := renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(NewGemRenderer(config), 1000)),
	)
	return r
}

// A GemRenderer struct is an implementation of renderer.GemRenderer that renders
// nodes as gemtext.
type GemRenderer struct {
	config Config
}

// NewGemRenderer returns a new renderer.NodeRenderer.
func NewGemRenderer(config *Config) *GemRenderer {
	r := &GemRenderer{
		config: *config,
	}
	return r
}

// gem must implement renderer.NodeRenderer
var _ renderer.NodeRenderer = &GemRenderer{}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *GemRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

	// extras
	reg.Register(east.KindStrikethrough, r.renderStrikethrough)
}

// linkOnly is a helper function that returns true is a node's subnodes have
// links and don't have text. This is used for checking if a heading/paragraph
// is actually JUST a link.
func linkOnly(source []byte, node ast.Node) bool {
	var hasLink bool = false
	var hasText bool = false
	// check if the paragraph contains ONLY links
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch nl := child.(type) {
		case *ast.Link:
			hasLink = true
		case *ast.AutoLink:
			hasLink = true
		case *ast.Text:
			if string(nl.Segment.Value(source)) != "" {
				hasText = true
			}
		}
	}
	if hasLink && !hasText {
		return true
	}
	return false
}

func linkText(source *[]byte, l *ast.Link) ([]byte, error) {
	var buf bytes.Buffer
	for child := l.FirstChild(); child != nil; child = child.NextSibling() {
		sub := New()
		if err := sub.Render(&buf, *source, child); err != nil {
			return nil, err
		}
	}
	text := bytes.TrimSpace(buf.Bytes())
	return text, nil
}
