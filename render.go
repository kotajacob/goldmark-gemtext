package gemtext

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yuin/goldmark/ast"
)

var (
	Logger = log.New(os.Stderr, "", 0)
)

// Render write node as Markdown o writer.
func Render(w io.Writer, source []byte, node ast.Node) (err error) {
	defer func() {
		if p := recover(); p != nil && err == nil {
			if e, ok := p.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", p)
			}
		}
	}()

	write := func(str string, a ...interface{}) {
		if _, err = fmt.Fprintf(w, str, a...); err != nil {
			panic(err)
		}
	}

	return ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := node.(type) {
		case *ast.Document:
			// could do stuff with markdown metadata

		case *ast.Heading:
			if entering {
				switch n.Level {
				case 1:
					write("# ")
				case 2:
					write("## ")
				default:
					write("### ")
				}
			} else {
				write("\n\n")
			}

		case *ast.Blockquote:
			if entering {
				var buf bytes.Buffer
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					if err = Render(&buf, source, child); err != nil {
						return ast.WalkStop, err
					}
				}

				text := bytes.TrimSpace(buf.Bytes())
				lines := bytes.SplitAfter(text, []byte{'\n'})
				for _, line := range lines {
					write(">")
					if len(line) > 0 && line[0] != '>' && line[0] != '\n' {
						write(" ")
					}
					write("%s", line)
				}

				return ast.WalkSkipChildren, nil
			} else {
				write("\n\n")
			}

		case *ast.CodeBlock:
			if entering {
				write("```")
				write("\n")
				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					write("%s", line.Value(source))
				}

				write("```")
				return ast.WalkSkipChildren, nil
			} else {
				write("\n\n")
			}

		case *ast.FencedCodeBlock:
			if entering {
				write("```")
				if n.Info != nil {
					write("%s", n.Info.Segment.Value(source))
				}
				write("\n")

				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					write("%s", line.Value(source))
				}

				write("```")
				return ast.WalkSkipChildren, nil
			} else {
				write("\n\n")
			}

		case *ast.HTMLBlock:
			return ast.WalkSkipChildren, nil

		case *ast.List:
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
						if err = Render(&buf, source, chld); err != nil {
							return ast.WalkStop, err
						}
					}

					// print list item
					write("* ")

					text := bytes.TrimSpace(buf.Bytes())
					buf.Reset()

					lines := bytes.SplitAfter(text, []byte{'\n'})
					for i, line := range lines {
						if i > 0 && len(line) > 0 && line[0] != '\n' {
							write(indent)
						}
						write("%s", line)
					}

					write("\n")
					if !n.IsTight {
						write("\n")
					}
				}

				if n.IsTight {
					write("\n")
				}

				return ast.WalkSkipChildren, nil
			}

		case *ast.ListItem:
			// return ast.WalkSkipChildren, nil

		case *ast.Paragraph:
			if !entering {
				firstLink := true
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					switch nl := child.(type) {
					case *ast.Link:
						if firstLink {
							write("\n")
						}
						var buf bytes.Buffer
						for chld := nl.FirstChild(); chld != nil; chld = chld.NextSibling() {
							if err = Render(&buf, source, chld); err != nil {
								return ast.WalkStop, err
							}
						}
						text := bytes.TrimSpace(buf.Bytes())
						buf.Reset()
						write("\n=> %s %s", nl.Destination, text)
						firstLink = false
					case *ast.AutoLink:
						if firstLink {
							write("\n")
						}
						write("\n=> %s", nl.Label(source))
						firstLink = false
					}
				}
				write("\n\n")
			}

		case *ast.TextBlock:
			if !entering {
				if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
					write("\n")
				}
			}

		case *ast.ThematicBreak:
			if entering {
				for i := 0; i < 80; i++ {
					write("-")
				}
			} else {
				write("\n\n")
			}

		case *ast.AutoLink:
			if entering {
				write("%s", n.Label(source))
			}

		case *ast.CodeSpan:
			// hide symbols

		case *ast.Emphasis:
			// hide symbols

		case *ast.Link:
			// hide symbols

		case *ast.Image:
			if entering {
				write("=> ")
				write("%s ", n.Destination)
			} else {
				if n.Type() == ast.TypeInline {
					write("\n")
				}
			}

		case *ast.RawHTML:
			// skip
			return ast.WalkSkipChildren, nil

		case *ast.Text:
			if entering {
				write("%s", n.Segment.Value(source))
				if n.SoftLineBreak() {
					if n.NextSibling().Kind() != ast.KindImage {
						write(" ")
					}
				}
			}

		case *ast.String:
			if entering {
				write("%s", n.Value)
			}

		default:
			if Logger != nil && entering {
				Logger.Printf("WARNING: unsupported AST %v type", node.Kind())
			}
		}
		return ast.WalkContinue, nil
	})
}
