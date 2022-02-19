package gemtext

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func ExampleNew() {
	src := `
# This is a heading

This is a [paragraph](https://en.wikipedia.org/wiki/Paragraph) with [some
links](https://en.wikipedia.org/wiki/Hyperlink) in it.

Next we'll have a list of some musicians I like, but as an individual list of
links. One of the neat features of goldmark-gemtext is that it recognizes when
a "paragraph" is really just a list of links and handles it as if it's a list
of links by simply converting them to the gemtext format. I wasn't able to find
any other markdown to gemtext tools that could do this so it was the
inspiration for writing this in the first place.

[Noname](https://nonameraps.bandcamp.com/)\
[Milo](https://afrolab9000.bandcamp.com/album/so-the-flies-dont-come)\
[Busdriver](https://busdriver-thumbs.bandcamp.com/)\
[Neat Beats](https://www.youtube.com/watch?v=X6kGg31G0As)\
[Ratatat](http://www.ratatatmusic.com/)\
[Sylvan Esso](https://www.sylvanesso.com/)\
[Phoebe Bridgers](https://phoebefuckingbridgers.com/)
`
	// create markdown parser
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)

	// set some options
	options := []Option{WithHeadingLink(HeadingLinkAuto), WithCodeSpan(CodeSpanMarkdown)}

	md.SetRenderer(New(options...))
	_ = md.Convert([]byte(src), &buf) // ignoring errors for example
	fmt.Println(buf.String())
	// Output:
	// # This is a heading
	//
	// This is a paragraph with some links in it.
	//
	// => https://en.wikipedia.org/wiki/Paragraph paragraph
	// => https://en.wikipedia.org/wiki/Hyperlink some links
	//
	// Next we'll have a list of some musicians I like, but as an individual list of links. One of the neat features of goldmark-gemtext is that it recognizes when a "paragraph" is really just a list of links and handles it as if it's a list of links by simply converting them to the gemtext format. I wasn't able to find any other markdown to gemtext tools that could do this so it was the inspiration for writing this in the first place.
	//
	// => https://nonameraps.bandcamp.com/ Noname
	// => https://afrolab9000.bandcamp.com/album/so-the-flies-dont-come Milo
	// => https://busdriver-thumbs.bandcamp.com/ Busdriver
	// => https://www.youtube.com/watch?v=X6kGg31G0As Neat Beats
	// => http://www.ratatatmusic.com/ Ratatat
	// => https://www.sylvanesso.com/ Sylvan Esso
	// => https://phoebefuckingbridgers.com/ Phoebe Bridgers
}
