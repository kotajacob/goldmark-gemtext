package gemtext

import (
	"fmt"

	"git.sr.ht/~kota/fuckery"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/util"
)

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
