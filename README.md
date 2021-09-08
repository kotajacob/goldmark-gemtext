# goldmark-gemtext [![godocs.io](https://godocs.io/git.sr.ht/~kota/goldmark-gemtext?status.svg)](https://godocs.io/git.sr.ht/~kota/goldmark-gemtext) [![builds.sr.ht status](https://builds.sr.ht/~kota/goldmark-gemtext.svg)](https://builds.sr.ht/~kota/goldmark-gemtext)

A gemtext renderer for [goldmark](https://github.com/yuin/goldmark). You can use
this library to parse commonmark markdown (with additional support for autolinks
and strikethrough) and convert it into nice clean looking gemtext with a few
options. This library is used to do the heavy lifting in
[gemgen](https://git.sr.ht/~kota/gemgen).

You create a renderer with New(option...) and pass in options:
```go
var src = `
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

// Set some options.
var options = []Option{WithHeadingLink(HeadingLinkAuto), WithCodeSpan(CodeSpanMarkdown)}

md.SetRenderer(New(options...))
_ = md.Convert([]byte(src), &buf) // ignoring errors for example
fmt.Println(buf.String())
```
