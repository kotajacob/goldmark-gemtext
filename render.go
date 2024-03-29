package gemtext

import (
	"bytes"
	"fmt"
	"io"

	wast "git.sr.ht/~kota/goldmark-wiki/ast"
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
		renderer.WithNodeRenderers(
			util.Prioritized(NewGemRenderer(config), 1000),
		),
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
	reg.Register(wast.KindWiki, r.renderWiki)
}

// linkOnly is a helper function that returns true is a node's subnodes have
// links and don't have text. This is used for checking if a heading/paragraph
// is actually JUST a link.
func linkOnly(source []byte, node ast.Node) bool {
	var hasLink bool = false
	var hasText bool = false
	// Check if the paragraph contains ONLY links.
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch nl := child.(type) {
		case *ast.Link:
			hasLink = true
		case *ast.AutoLink:
			hasLink = true
		case *wast.Wiki:
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

// linkPrint is a helper function that prints a link's text to a writer, applies
// any regex replacers. Images are not handled by this function as they operate
// slightly differently. Format can be used to format the link text.
// Returns false if a link was not printed.
func linkPrint(w io.Writer, source []byte, node ast.Node, replacers []LinkReplacer, format string) bool {
	if format == "" {
		format = "%s"
	}
	// I know the logic is nearly duplicated in *ast.Link and *wast.Wiki, but I
	// don't know of a good way to consolidate this. You _can_ match multiple
	// types in a type switch, but instead of n being the correct type it will
	// be assigned interface{}. So you would need to do additional type
	// assertions and that seems worse than a little duplication.
	switch n := node.(type) {
	case *ast.Link:
		// Apply link replacers.
		destination := n.Destination
		for _, r := range replacers {
			s := r.replace(string(destination), LinkMarkdown)
			destination = []byte(s)
		}

		// Get link text.
		text, err := nodeText(source, n)
		if err != nil {
			return false
		}
		fmt.Fprintf(w, "=> %s %s", destination, fmt.Sprintf(format, text))
		return true
	case *wast.Wiki:
		// Apply link replacers.
		destination := n.Destination
		for _, r := range replacers {
			s := r.replace(string(destination), LinkWiki)
			destination = []byte(s)
		}

		// Get link text.
		text, err := nodeText(source, n)
		if err != nil {
			return false
		}

		fmt.Fprintf(w, "=> %s %s", destination, fmt.Sprintf(format, text))
		return true
	case *ast.AutoLink:
		// Apply link replacers.
		destination := n.Label(source)
		for _, r := range replacers {
			s := r.replace(string(destination), LinkAuto)
			destination = []byte(s)
		}
		fmt.Fprintf(w, "=> %s", destination)
		return true
	}
	return false
}

// replace applies a LinkReplacer if the type matches t. The string returned
// will be modified if it matched.
func (r LinkReplacer) replace(s string, t LinkType) string {
	if t == r.Type {
		return r.Regex.ReplaceAllString(s, r.Replacement)
	}
	return s
}

// nodeText is a helper function that recursively creates and runs a renderer
// for a specific node. This is slower, but is the only way to handle some link
// text edge cases (multiline links, emphasis markings in link test, etc).
func nodeText(source []byte, node ast.Node) ([]byte, error) {
	var buf bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		sub := New()
		if err := sub.Render(&buf, source, child); err != nil {
			return nil, err
		}
	}
	text := bytes.TrimSpace(buf.Bytes())
	return text, nil
}
