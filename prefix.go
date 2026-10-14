package main

import (
	"bytes"
	"io"
)

// linePrefixWriter labels each output line even when writes split a line.
type linePrefixWriter struct {
	w      io.Writer
	prefix []byte
	start  bool
}

func newLinePrefixWriter(w io.Writer, prefix string) io.Writer {
	return &linePrefixWriter{w: w, prefix: []byte(prefix), start: true}
}

func (w *linePrefixWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		if w.start {
			if n, err := w.w.Write(w.prefix); err != nil {
				return written, err
			} else if n != len(w.prefix) {
				return written, io.ErrShortWrite
			}
			w.start = false
		}
		end := bytes.IndexByte(p, '\n')
		if end < 0 {
			end = len(p)
		} else {
			end++
		}
		n, err := w.w.Write(p[:end])
		written += n
		if err != nil {
			return written, err
		}
		if n != end {
			return written, io.ErrShortWrite
		}
		w.start = p[end-1] == '\n'
		p = p[end:]
	}
	return written, nil
}
