package gemtext

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRender(t *testing.T) {
	source, err := os.ReadFile("test_data/render.md")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_data/render.gmi")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer

	HeadingLinks = true
	Emphasis = false
	UnicodeEmphasis = false
	CodeSpan = false
	Strikethrough = false
	err = Format(source, &buf)
	if err != nil {
		t.Fatal(err)
	}

	got := buf.Bytes()
	if !cmp.Equal(got, want) {
		fmt.Println(cmp.Diff(got, want))
	}

}
