package main

import (
	"bytes"
	"testing"
)

func TestLinePrefixWriterAcrossWrites(t *testing.T) {
	var out bytes.Buffer
	w := newLinePrefixWriter(&out, "source.json: ")
	for _, chunk := range []string{"one", "\ntwo\nthree", "\nfour\n"} {
		if _, err := w.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	want := "source.json: one\nsource.json: two\nsource.json: three\nsource.json: four\n"
	if out.String() != want {
		t.Fatalf("want %q; got %q", want, out.String())
	}
}
