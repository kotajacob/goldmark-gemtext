package gemtext

import (
	"bytes"
	"fmt"

	wast "git.sr.ht/~kota/goldmark-wiki/ast"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

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
					// Print the first link we find, then skip then rest of the node.
					case *ast.Link:
						fmt.Fprintf(w, "=> %s %s", nl.Destination, nl.Text(source))
						return ast.WalkSkipChildren, nil
					case *wast.Wiki:
						fmt.Fprintf(w, "=> %s %s", nl.Destination, nl.Text(source))
						return ast.WalkSkipChildren, nil
					case *ast.AutoLink:
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
					case *wast.Wiki:
						fmt.Fprintf(w, "%s", nl.Text(source))
					case *ast.AutoLink:
						fmt.Fprintf(w, "%s", nl.Label(source))
					}
				}
			}
		}
	} else {
		if r.config.HeadingSpace == HeadingSpaceSingle {
			fmt.Fprintf(w, "\n")
		} else {
			fmt.Fprintf(w, "\n\n")
		}
		if r.config.HeadingLink == HeadingLinkBelow {
			// Print all links that were in the heading below the heading.
			var hasLink bool
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				switch nl := child.(type) {
				case *ast.Link:
					hasLink = true
					fmt.Fprintf(w, "=> %s %s\n", nl.Destination, nl.Text(source))
				case *wast.Wiki:
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
					fmt.Fprint(w, indent)
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

// renderParagraphLinkOnly is called is used to print a paragraph which
// contains links (auto or normal) and no text. The linkOnly helper function is
// used to test this condition. Link only paragraphs are simply renderered as a
// list of gemini links.
func (r *GemRenderer) renderParagraphLinkOnly(w util.BufWriter, source []byte, n *ast.Paragraph, entering bool) (ast.WalkStatus, error) {
	firstLink := true
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch nl := child.(type) {
		case *ast.Link:
			if !firstLink {
				// add line breaks between links
				fmt.Fprintf(w, "\n")
			}
			text, err := nodeText(&source, nl)
			if err != nil {
				return ast.WalkStop, nil
			}
			fmt.Fprintf(w, "=> %s %s", nl.Destination, text)
			firstLink = false
		case *wast.Wiki:
			if !firstLink {
				// add line breaks between links
				fmt.Fprintf(w, "\n")
			}
			text, err := nodeText(&source, nl)
			if err != nil {
				return ast.WalkStop, nil
			}
			fmt.Fprintf(w, "=> %s %s", nl.Destination, text)
			firstLink = false
		case *ast.AutoLink:
			if !firstLink {
				// add line breaks between links
				fmt.Fprintf(w, "\n")
			}
			fmt.Fprintf(w, "=> %s", nl.Label(source))
			firstLink = false
		}
	}
	fmt.Fprintf(w, "\n\n")
	return ast.WalkContinue, nil
}

// renderParagraphLinkOff renders the paragraph without printing links below.
// If the paragraph is "link only" it will print itself as a link since it
// shouldn't really be considered a paragraph.
func (r *GemRenderer) renderParagraphLinkOff(w util.BufWriter, source []byte, n *ast.Paragraph, entering bool) (ast.WalkStatus, error) {
	if !entering {
		// We can make this check inside !entering, because link only
		// paragraphs do not contain text. It's a weird quick of goldmark and
		// this is the work-around.
		if linkOnly(source, n) {
			return r.renderParagraphLinkOnly(w, source, n, entering)
		}
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

// renderParagraphLinkBelow is the default paragraph printing mode. Links are
// printed below the paragraph in a list. If the paragraph contains only links
// it is printed as a link or list of links itself.
func (r *GemRenderer) renderParagraphLinkBelow(w util.BufWriter, source []byte, n *ast.Paragraph, entering bool) (ast.WalkStatus, error) {
	if !entering {
		// We can make this check inside !entering, because link only
		// paragraphs do not contain text. It's a weird quick of goldmark and
		// this is the work-around.
		if linkOnly(source, n) {
			return r.renderParagraphLinkOnly(w, source, n, entering)
		}
		// Handle links in non-link-only paragraphs.
		firstLink := true
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch nl := child.(type) {
			case *ast.Link:
				if firstLink {
					fmt.Fprintf(w, "\n")
				}
				text, err := nodeText(&source, nl)
				if err != nil {
					return ast.WalkStop, nil
				}
				fmt.Fprintf(w, "\n")
				fmt.Fprintf(w, "=> %s %s", nl.Destination, text)
				firstLink = false
			case *wast.Wiki:
				if firstLink {
					fmt.Fprintf(w, "\n")
				}
				text, err := nodeText(&source, nl)
				if err != nil {
					return ast.WalkStop, nil
				}
				fmt.Fprintf(w, "\n")
				fmt.Fprintf(w, "=> %s %s", nl.Destination, text)
				firstLink = false
			case *ast.AutoLink:
				if firstLink {
					fmt.Fprintf(w, "\n\n")
				}
				fmt.Fprintf(w, "=> %s", nl.Label(source))
				firstLink = false
			}
		}
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Paragraph)
	switch r.config.ParagraphLink {
	case ParagraphLinkOff:
		status, err := r.renderParagraphLinkOff(w, source, n, entering)
		return status, err
	default:
		status, err := r.renderParagraphLinkBelow(w, source, n, entering)
		return status, err
	}
}

func (r *GemRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.TextBlock)
	if !entering {
		if n.NextSibling() != nil && n.FirstChild() != nil {
			fmt.Fprintf(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *GemRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// TODO: make this configurable
	if entering {
		fmt.Fprintf(w, r.config.HorizontalRule)
	} else {
		fmt.Fprintf(w, "\n\n")
	}
	return ast.WalkContinue, nil
}
