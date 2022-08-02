package gemtext

import (
	"fmt"

	"git.sr.ht/~kota/fuckery"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/util"
)

// renderStrikethrough writes strikethrough text based on a few config options.
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

// renderWiki writes a wiki style link in gemtext.
// Similar to links and autolinks the node is skipped if the parent node
// contains only links. We use the parent node (paragraph or heading) to do the
// actual heavy lifting.
func (r *GemRenderer) renderWiki(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// skip if the parent node contains only links
	if linkOnly(source, node.Parent()) {
		return ast.WalkSkipChildren, nil
	}
	curly := r.config.ParagraphLink == ParagraphLinkCurlyBelow
	if entering {
		if curly {
			fmt.Fprint(w, "{")
		}
	} else {
		if curly {
			fmt.Fprint(w, "}")
		}
	}

	return ast.WalkContinue, nil
}
