package gemtext

import (
	"bytes"
	"fmt"

	"git.sr.ht/~kota/fuckery"
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
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
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
	if hasLink == true && hasText == false {
		return true
	}
	return false
}

func (r *GemRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		// Check if the heading contains only links.
		if r.config.HeadingLink == HeadingLinkAuto {
			if linkOnly(source, n) {
				// In Auto mode, link only headings prints their first link then exit.
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					switch nl := child.(type) {
					case *ast.Link:
						// Print the first link we find, then skip then rest of the node.
						fmt.Fprintf(w, "=> %s %s", nl.Destination, nl.Text(source))
						return ast.WalkSkipChildren, nil
					case *ast.AutoLink:
						// Print the first link we find, then skip then rest of the node.
						fmt.Fprintf(w, "=> %s", nl.Label(source))
						return ast.WalkSkipChildren, nil
					}
				}
			}
		}

		// Print the heading. Automode link only headings wont make it this far.
		switch n.Level {
		case 1:
			fmt.Fprintf(w, "# ")
		case 2:
			fmt.Fprintf(w, "## ")
		default:
			fmt.Fprintf(w, "### ")
		}

		if r.config.HeadingLink == HeadingLinkOff || r.config.HeadingLink == HeadingLinkBelow {
			// Check if it's link only to print the link labels. Link labels
			// are not printed by links if their parent is link only.
			if linkOnly(source, n) {
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					switch nl := child.(type) {
					case *ast.Link:
						fmt.Fprintf(w, "%s", nl.Text(source))
					case *ast.AutoLink:
						fmt.Fprintf(w, "%s", nl.Label(source))
					}
				}
			}
		}
	} else {
		fmt.Fprintf(w, "\n\n")
		if r.config.HeadingLink == HeadingLinkBelow {
			// Print all links that were in the heading below the heading.
			var hasLink bool
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				switch nl := child.(type) {
				case *ast.Link:
					hasLink = true
					fmt.Fprintf(w, "=> %s %s\n", nl.Destination, nl.Text(source))
				case *ast.AutoLink:
					hasLink = true
					fmt.Fprintf(w, "=> %s\n", nl.Label(source))
				}
			}
			if hasLink {
				// Print an extra newline after the last link, if the heading
				// had links.
				fmt.Fprintf(w, "\n")
			}
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Blockquote)
	if entering {
		var buf bytes.Buffer
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			sub := New()
			if err := sub.Render(&buf, source, child); err != nil {
				return ast.WalkStop, err
			}
		}

		text := bytes.TrimSpace(buf.Bytes())
		lines := bytes.SplitAfter(text, []byte{'\n'})
		for _, line := range lines {
			fmt.Fprintf(w, ">")
			if len(line) > 0 && line[0] != '>' && line[0] != '\n' {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprintf(w, "%s", line)
		}

		return ast.WalkSkipChildren, nil
	} else {
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// NOTE: This differs slightly from FencedCodeBlock as it cannot contain an
	// info line.
	n := node.(*ast.CodeBlock)
	if entering {
		fmt.Fprintf(w, "```")
		fmt.Fprintf(w, "\n")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			fmt.Fprintf(w, "%s", line.Value(source))
		}

		fmt.Fprintf(w, "```")
		return ast.WalkSkipChildren, nil
	} else {
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		fmt.Fprintf(w, "```")
		if n.Info != nil {
			fmt.Fprintf(w, "%s", n.Info.Segment.Value(source))
		}
		fmt.Fprintf(w, "\n")

		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			fmt.Fprintf(w, "%s", line.Value(source))
		}

		fmt.Fprintf(w, "```")
		return ast.WalkSkipChildren, nil
	} else {
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip html block - can't be used
	return ast.WalkSkipChildren, nil
}

func (r *GemRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	if entering {
		start := n.Start
		if start == 0 {
			start = 1
		}
		indent := "  "

		var buf bytes.Buffer
		// all ListItems
		for nl := n.FirstChild(); nl != nil; nl = nl.NextSibling() {
			for chld := nl.FirstChild(); chld != nil; chld = chld.NextSibling() {
				sub := New()
				if err := sub.Render(&buf, source, chld); err != nil {
					return ast.WalkStop, err
				}
			}

			// print list item
			fmt.Fprintf(w, "* ")

			text := bytes.TrimSpace(buf.Bytes())
			buf.Reset()

			lines := bytes.SplitAfter(text, []byte{'\n'})
			for i, line := range lines {
				if i > 0 && len(line) > 0 && line[0] != '\n' {
					fmt.Fprintf(w, indent)
				}
				fmt.Fprintf(w, "%s", line)
			}

			fmt.Fprintf(w, "\n")
			if !n.IsTight {
				fmt.Fprintf(w, "\n")
			}
		}

		if n.IsTight {
			fmt.Fprintf(w, "\n")
		}

		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Paragraph)
	if !entering {
		// loop through links and place them outside paragraph
		firstLink := true
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch nl := child.(type) {
			case *ast.Link:
				if linkOnly(source, n) {
					if !firstLink {
						// add line breaks between links in a link only paragraph
						fmt.Fprintf(w, "\n")
					}
				} else {
					if firstLink {
						fmt.Fprintf(w, "\n")
					}
				}
				var buf bytes.Buffer
				// TODO: Can I just use nl.Text(source) instead of this shit?
				for chld := nl.FirstChild(); chld != nil; chld = chld.NextSibling() {
					sub := New()
					if err := sub.Render(&buf, source, chld); err != nil {
						return ast.WalkStop, err
					}
				}
				text := bytes.TrimSpace(buf.Bytes())
				buf.Reset()
				if !linkOnly(source, n) {
					fmt.Fprintf(w, "\n")
				}
				fmt.Fprintf(w, "=> %s %s", nl.Destination, text)
				firstLink = false
			case *ast.AutoLink:
				if linkOnly(source, n) {
					if !firstLink {
						fmt.Fprintf(w, "\n")
					}
				} else {
					if firstLink {
						fmt.Fprintf(w, "\n\n")
					}
				}
				fmt.Fprintf(w, "=> %s", nl.Label(source))
				firstLink = false
			}
		}
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.TextBlock)
	if !entering {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			fmt.Fprintf(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// TODO: make this configurable
	if entering {
		for i := 0; i < 80; i++ {
			fmt.Fprintf(w, "-")
		}
	} else {
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip if the parent node contains only links
	n := node.(*ast.AutoLink)
	if entering {
		if linkOnly(source, node.Parent()) {
			return ast.WalkSkipChildren, nil
		} else {
			fmt.Fprintf(w, string(n.Label(source)))
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch r.config.CodeSpan {
	case CodeSpanMarkdown:
		fmt.Fprintf(w, "`")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	if entering {
		switch r.config.Emphasis {
		case EmphasisMarkdown:
			if n.Level == 1 {
				fmt.Fprintf(w, "_")
			} else {
				fmt.Fprintf(w, "**")
			}
		case EmphasisUnicode:
			if n.Level == 1 {
				fmt.Fprintf(w, "%s", fuckery.ItalicSans(string(n.Text(source))))
				return ast.WalkSkipChildren, nil
			} else {
				fmt.Fprintf(w, "%s", fuckery.BoldSans(string(n.Text(source))))
				return ast.WalkSkipChildren, nil
			}
		}
	} else {
		switch r.config.Emphasis {
		case EmphasisMarkdown:
			if n.Level == 1 {
				fmt.Fprintf(w, "_")
			} else {
				fmt.Fprintf(w, "**")
			}
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Image)
	if entering {
		fmt.Fprintf(w, "=> ")
		fmt.Fprintf(w, "%s ", n.Destination)
	} else {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			fmt.Fprintf(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip if the parent node contains only links
	if linkOnly(source, node.Parent()) {
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip raw html - can't be used
	return ast.WalkSkipChildren, nil
}

func (r *GemRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip if the parent node contains only links
	if linkOnly(source, node.Parent()) {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.Text)
	if entering {
		fmt.Fprintf(w, "%s", n.Segment.Value(source))
		// use a space for soft line breaks unless the next node is an image
		if n.SoftLineBreak() {
			if n.NextSibling().Kind() != ast.KindImage {
				fmt.Fprintf(w, " ")
			}
		}
		if n.HardLineBreak() {
			fmt.Fprintf(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	fmt.Fprintf(w, "%s", n.Value)
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderStrikethrough(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*east.Strikethrough)
	if entering {
		switch r.config.Strikethrough {
		case StrikethroughMarkdown:
			fmt.Fprintf(w, "~~")
		case StrikethroughUnicode:
			fmt.Fprintf(w, "%s", fuckery.Strike(string(n.Text(source))))
			return ast.WalkSkipChildren, nil
		}
	} else {
		switch r.config.Strikethrough {
		case StrikethroughMarkdown:
			fmt.Fprintf(w, "~~")
		}
	}
	return ast.WalkContinue, nil
}
