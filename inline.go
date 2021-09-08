package gemtext

import (
	"fmt"

	"git.sr.ht/~kota/fuckery"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

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
