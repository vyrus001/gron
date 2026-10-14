package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestStreamHandlesAdjacentValuesAndLongLine(t *testing.T) {
	long := strings.Repeat("x", 1024*1024+1)
	in := strings.NewReader(`{"value":"` + long + `"}{"value":2}`)
	var out bytes.Buffer
	code, err := gronStream(in, &out, optMonochrome)
	if code != exitOK || err != nil {
		t.Fatalf("gronStream: code=%d err=%v", code, err)
	}
	if !strings.Contains(out.String(), `json[0].value = "`+long+`";`) || !strings.Contains(out.String(), "json[1].value = 2;") {
		t.Fatal("stream output missing either value")
	}
}

func TestGronRequiresStreamForMultipleValues(t *testing.T) {
	for _, opts := range []int{optMonochrome, optMonochrome | optNoSort} {
		var out bytes.Buffer
		code, err := gron(strings.NewReader(`{"a":1}{"b":2}`), &out, opts)
		if code == exitOK || err == nil || !strings.Contains(err.Error(), "use --stream") {
			t.Fatalf("opts=%d code=%d err=%v", opts, code, err)
		}
	}
}

func TestNoSortKeepsInputObjectOrder(t *testing.T) {
	for _, action := range []actionFn{gron, gronStream} {
		var out bytes.Buffer
		code, err := action(strings.NewReader(`{"z":1,"a":2}`), &out, optMonochrome|optNoSort)
		if code != exitOK || err != nil {
			t.Fatalf("no-sort: code=%d err=%v", code, err)
		}
		text := out.String()
		if strings.Index(text, ".z = 1;") < 0 || strings.Index(text, ".z = 1;") > strings.Index(text, ".a = 2;") {
			t.Fatalf("input order lost: %q", text)
		}
	}
}

func TestUngronStreamRestoresJSONL(t *testing.T) {
	var gronned, restored bytes.Buffer
	code, err := gronStream(strings.NewReader("{\"a\":1}\n{\"b\":2}\n"), &gronned, optMonochrome)
	if code != exitOK || err != nil {
		t.Fatalf("gronStream: code=%d err=%v", code, err)
	}
	code, err = ungron(&gronned, &restored, optMonochrome|optStream)
	if code != exitOK || err != nil || restored.String() != "{\"a\":1}\n{\"b\":2}\n" {
		t.Fatalf("ungron stream: code=%d err=%v output=%q", code, err, restored.String())
	}
}

func TestUngronHandlesLongLineAndTypeReplacement(t *testing.T) {
	long := strings.Repeat("x", 1024*1024+1)
	for _, input := range []string{
		`json.a = {};` + "\n" + `json.a = "` + long + `";`,
		`json.a = "old";` + "\n" + `json.a = {};` + "\n" + `json.a.value = "` + long + `";`,
	} {
		var out bytes.Buffer
		code, err := ungron(strings.NewReader(input), &out, optMonochrome)
		if code != exitOK || err != nil {
			t.Fatalf("ungron: code=%d err=%v", code, err)
		}
		if !strings.Contains(out.String(), long) {
			t.Fatal("ungron output missing long value")
		}
	}
}

func TestGronValuesUsesProvidedStreams(t *testing.T) {
	var out bytes.Buffer
	code, err := gronValues(strings.NewReader(`json.value = "hello";`), &out, 0)
	if code != exitOK || err != nil || out.String() != "hello\n" {
		t.Fatalf("gronValues: code=%d err=%v output=%q", code, err, out.String())
	}
}

func TestGronAcceptsUTF8BOM(t *testing.T) {
	for _, action := range []actionFn{gron, gronStream} {
		var out bytes.Buffer
		code, err := action(strings.NewReader("\xef\xbb\xbf"+`{"value":1}`), &out, optMonochrome)
		if code != exitOK || err != nil || !strings.Contains(out.String(), ".value = 1;") {
			t.Fatalf("BOM input: code=%d err=%v output=%q", code, err, out.String())
		}
	}
}

func TestUngronAcceptsColoredStatements(t *testing.T) {
	in := "\x1b[31mjson.value\x1b[0m = \x1b[1;33m\"hello\"\x1b[0m;\n"
	var out bytes.Buffer
	code, err := ungron(strings.NewReader(in), &out, optMonochrome)
	if code != exitOK || err != nil || !strings.Contains(out.String(), `"value": "hello"`) {
		t.Fatalf("colored input: code=%d err=%v output=%q", code, err, out.String())
	}
}

func TestGron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.gron"},
		{"testdata/two.json", "testdata/two.gron"},
		{"testdata/three.json", "testdata/three.gron"},
		{"testdata/github.json", "testdata/github.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gron(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.gron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestLargeGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/long-stream.json", "testdata/long-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestUngron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.gron", "testdata/one.json"},
		{"testdata/two.gron", "testdata/two.json"},
		{"testdata/three.gron", "testdata/three.json"},
		{"testdata/grep-separators.gron", "testdata/grep-separators.json"},
		{"testdata/github.gron", "testdata/github.json"},
		{"testdata/large-line.gron", "testdata/large-line.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := ungron(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}

func TestGronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.jgron"},
		{"testdata/two.json", "testdata/two.jgron"},
		{"testdata/three.json", "testdata/three.jgron"},
		{"testdata/github.json", "testdata/github.jgron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gron(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestGronStreamJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.jgron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.jgron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestUngronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.jgron", "testdata/one.json"},
		{"testdata/two.jgron", "testdata/two.json"},
		{"testdata/three.jgron", "testdata/three.json"},
		{"testdata/github.jgron", "testdata/github.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := ungron(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}

func BenchmarkBigJSON(b *testing.B) {
	in, err := os.Open("testdata/big.json")
	if err != nil {
		b.Fatalf("failed to open test data file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		out := &bytes.Buffer{}
		_, err = in.Seek(0, 0)
		if err != nil {
			b.Fatalf("failed to rewind input: %s", err)
		}

		_, err := gron(in, out, optMonochrome|optNoSort)
		if err != nil {
			b.Fatalf("failed to gron: %s", err)
		}
	}
}
