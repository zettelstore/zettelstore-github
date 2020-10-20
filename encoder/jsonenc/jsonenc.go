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

// Package jsonenc encodes the abstract syntax tree into JSON.
package jsonenc

import (
	"bytes"
	"io"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
)

func init() {
	encoder.Register("json", encoder.Info{
		Create:  func() encoder.Encoder { return &jsonEncoder{} },
		Default: true,
	})
}

type jsonEncoder struct{}

// SetOption sets an option for the encoder
func (je *jsonEncoder) SetOption(option encoder.Option) {}

// WriteZettel writes the encoded zettel to the writer.
func (je *jsonEncoder) WriteZettel(w io.Writer, zettel *ast.Zettel) (int, error) {
	b := encoder.NewBufWriter(w)
	b.WriteString("{\"meta\":")
	writeMeta(&b, zettel.Meta)
	b.WriteString(",\"content\":")
	writeEscaped(&b, zettel.Content.AsString())
	b.WriteByte('}')
	length, err := b.Flush()
	return length, err
}

// WriteMeta encodes meta data as HTML5.
func (je *jsonEncoder) WriteMeta(w io.Writer, meta *domain.Meta) (int, error) {
	b := encoder.NewBufWriter(w)
	writeMeta(&b, meta)
	length, err := b.Flush()
	return length, err
}

func (je *jsonEncoder) WriteContent(w io.Writer, zettel *ast.Zettel) (int, error) {
	b := encoder.NewBufWriter(w)
	writeEscaped(&b, zettel.Content.AsString())
	length, err := b.Flush()
	return length, err
}

// WriteBlocks writes a block slice to the writer
func (je *jsonEncoder) WriteBlocks(w io.Writer, bs ast.BlockSlice) (int, error) {
	return 0, encoder.ErrNoWriteBlocks
}

// WriteInlines writes an inline slice to the writer
func (je *jsonEncoder) WriteInlines(w io.Writer, is ast.InlineSlice) (int, error) {
	return 0, encoder.ErrNoWriteInlines
}

func writeMeta(b *encoder.BufWriter, meta *domain.Meta) {
	b.WriteByte('{')
	first := true
	for _, p := range meta.Pairs() {
		if !first {
			b.WriteString(",\"")
		} else {
			b.WriteByte('"')
			first = false
		}
		b.Write(Escape(p.Key))
		b.WriteString("\":\"")
		b.Write(Escape(p.Value))
		b.WriteByte('"')
	}
	b.WriteByte('}')
}

var (
	jsBackslash   = []byte{'\\', '\\'}
	jsDoubleQuote = []byte{'\\', '"'}
	jsNewline     = []byte{'\\', 'n'}
	jsTab         = []byte{'\\', 't'}
	jsCr          = []byte{'\\', 'r'}
	jsUnicode     = []byte{'\\', 'u', '0', '0', '0', '0'}
	jsHex         = []byte("0123456789ABCDEF")
)

// Escape returns the given string as a byte slice, where every non-printable
// rune is made printable.
func Escape(s string) []byte {
	var buf bytes.Buffer

	last := 0
	for i, ch := range s {
		var b []byte
		switch ch {
		case '\t':
			b = jsTab
		case '\r':
			b = jsCr
		case '\n':
			b = jsNewline
		case '"':
			b = jsDoubleQuote
		case '\\':
			b = jsBackslash
		default:
			if ch < ' ' {
				b = jsUnicode
				b[2] = '0'
				b[3] = '0'
				b[4] = jsHex[ch>>4]
				b[5] = jsHex[ch&0xF]
			} else {
				continue
			}
		}
		buf.WriteString(s[last:i])
		buf.Write(b)
		last = i + 1
	}
	buf.WriteString(s[last:])
	return buf.Bytes()
}

func writeEscaped(b *encoder.BufWriter, s string) {
	b.WriteByte('"')
	b.Write(Escape(s))
	b.WriteByte('"')
}
