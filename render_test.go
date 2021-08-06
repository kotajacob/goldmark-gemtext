package gemtext

import (
	"os"
	"testing"
)

func TestHeadingLinks(t *testing.T) {
	source := []byte("# [twitter](https://twitter.com)")

	HeadingLinks = false
	err := Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmphasis(t *testing.T) {
	source := []byte("This sentence should have _some_ **emphasis** in it.")

	Emphasis = true
	err := Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCodeSpan(t *testing.T) {
	source := []byte("This sentence should have `some codespan in` it.")

	CodeSpan = true
	err := Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrikethrough(t *testing.T) {
	source := []byte("This sentence should have ~~some strikethrough in~~ it.")

	Strikethrough = true
	err := Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatter(t *testing.T) {
	source, err := os.ReadFile("sample.md")
	if err != nil {
		t.Fatal(err)
	}

	HeadingLinks = true
	Emphasis = false
	CodeSpan = false
	Strikethrough = false
	err = Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}
