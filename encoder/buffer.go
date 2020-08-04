//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Zettelstore is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Zettelstore. If not, see <http://www.gnu.org/licenses/>.
//-----------------------------------------------------------------------------

// Package encoder provides a generic interface to encode the abstract syntax
// tree into some text form.
package encoder

import (
	"encoding/base64"
	"io"
)

// BufWriter is a specialized buffered writer for encoding zettel.
type BufWriter struct {
	w      io.Writer // The io.Writer to write to
	err    error     // Collect error
	length int       // Sum length
	buf    []byte    // Buffer to collect bytes
}

// NewBufWriter creates a new BufWriter
func NewBufWriter(w io.Writer) BufWriter {
	return BufWriter{w: w, buf: make([]byte, 0, 4096)}
}

// Write writes the contents of p into the buffer.
func (w *BufWriter) Write(p []byte) (int, error) {
	if w.err == nil {
		w.buf = append(w.buf, p...)
		if len(w.buf) > 2048 {
			w.flush()
		}
	}
	return len(p), nil
}

// WriteString writes the contents of s into the buffer.
func (w *BufWriter) WriteString(s string) {
	w.Write([]byte(s))
}

// WriteByte writes the content of b into the buffer.
func (w *BufWriter) WriteByte(b byte) error {
	w.buf = append(w.buf, b)
	return nil
}

// WriteBase64 writes the content of p into the buffer, encoded with base64.
func (w *BufWriter) WriteBase64(p []byte) {
	if w.err == nil {
		w.flush()
	}
	if w.err == nil {
		encoder := base64.NewEncoder(base64.StdEncoding, w.w)
		length, err := encoder.Write(p)
		w.length += length
		err1 := encoder.Close()
		if err == nil {
			w.err = err1
		} else {
			w.err = err
		}
	}
}

// Flush writes any buffered data to the underlying io.Writer. It returns the
// number of bytes written and an error if something went wrong.
func (w *BufWriter) Flush() (int, error) {
	if w.err == nil {
		w.flush()
	}
	return w.length, w.err
}

func (w *BufWriter) flush() {
	length, err := w.w.Write(w.buf)
	w.buf = w.buf[:0]
	w.length += length
	w.err = err
}
