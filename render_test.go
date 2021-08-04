package gemtext

import (
	"os"
	"testing"
)

func TestEmphasis(t *testing.T) {
	source := []byte("This sentence has _some_ **emphasis** in it.")

	Emphasis = true
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

	err = Format(source, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
}
