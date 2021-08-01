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

	isLinkOnly := func(n ast.Node) bool {
		var hasLink bool = false
		var hasText bool = false
		// check if the paragraph contains ONLY links
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
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
			// check if paragraph contains links and no text
			if !entering {
				// loop through links and place them outside paragraph
				firstLink := true
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					switch nl := child.(type) {
					case *ast.Link:
						if isLinkOnly(n) {
							if !firstLink {
								write("\n")
							}
						} else {
							if firstLink {
								write("\n")
							}
						}
						var buf bytes.Buffer
						for chld := nl.FirstChild(); chld != nil; chld = chld.NextSibling() {
							if err = Render(&buf, source, chld); err != nil {
								return ast.WalkStop, err
							}
						}
						text := bytes.TrimSpace(buf.Bytes())
						buf.Reset()
						if !isLinkOnly(n) {
							write("\n")
						}
						write("=> %s %s", nl.Destination, text)
						firstLink = false
					case *ast.AutoLink:
						if isLinkOnly(n) {
							if !firstLink {
								write("\n")
							}
						} else {
							if firstLink {
								write("\n\n")
							}
						}
						write("=> %s", nl.Label(source))
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
			// leave link as is in the text source
			if isLinkOnly(n.Parent()) {
				return ast.WalkSkipChildren, nil
			}
			if entering {
				write("%s", n.Label(source))
			}

		case *ast.CodeSpan:
			// hide symbols

		case *ast.Emphasis:
			// hide symbols

		case *ast.Link:
			if isLinkOnly(n.Parent()) {
				return ast.WalkSkipChildren, nil
			}

		case *ast.Image:
			if entering {
				write("=> ")
				write("%s ", n.Destination)
			} else {
				if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
					write("\n")
				}
			}

		case *ast.RawHTML:
			// skip
			return ast.WalkSkipChildren, nil

		case *ast.Text:
			if isLinkOnly(n.Parent()) {
				return ast.WalkSkipChildren, nil
			}
			if entering {
				write("%s", n.Segment.Value(source))
				if n.SoftLineBreak() {
					if n.NextSibling().Kind() != ast.KindImage {
						write(" ")
					}
				}
				if n.HardLineBreak() {
					write("\n")
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
