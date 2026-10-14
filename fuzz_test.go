package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func FuzzGronRoundTrip(f *testing.F) {
	for _, sample := range []string{`{}`, `[]`, `{"z":1,"a":[true,null]}`, `{"a b":"x","quote\"":"\n"}`} {
		f.Add([]byte(sample))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		if !json.Valid(input) {
			return
		}
		var statements bytes.Buffer
		if code, err := gron(bytes.NewReader(input), &statements, optMonochrome|optNoSort); code != exitOK || err != nil {
			t.Fatalf("gron: code=%d err=%v", code, err)
		}
		var output bytes.Buffer
		if code, err := ungron(&statements, &output, optMonochrome); code != exitOK || err != nil {
			t.Fatalf("ungron: code=%d err=%v", code, err)
		}
		var want, got interface{}
		wantDecoder := json.NewDecoder(bytes.NewReader(input))
		wantDecoder.UseNumber()
		if err := wantDecoder.Decode(&want); err != nil {
			t.Fatal(err)
		}
		gotDecoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
		gotDecoder.UseNumber()
		if err := gotDecoder.Decode(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatal("round trip changed JSON")
		}
	})
}
