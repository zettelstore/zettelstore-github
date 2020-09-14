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

	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
)

func acceptMeta(b *encoder.BufWriter, meta *domain.Meta, withTitle bool) {
	if withTitle {
		b.WriteString("\"title\":\"")
		b.Write(Escape(meta.GetDefault(domain.MetaKeyTitle, "")))
		b.WriteByte('"')
	}
	writeMetaList(b, meta, domain.MetaKeyTags, "tags")
	writeMetaString(b, meta, domain.MetaKeySyntax, "syntax")
	writeMetaString(b, meta, domain.MetaKeyRole, "role")
	if pairs := meta.PairsRest(); len(pairs) > 0 {
		b.WriteString(",\"header\":{\"")
		first := true
		for _, p := range pairs {
			if !first {
				b.WriteString("\",\"")
			}
			b.Write(Escape(p.Key))
			b.WriteString("\":\"")
			b.Write(Escape(p.Value))
			first = false
		}
		b.WriteString("\"}")
	}
}

func writeMetaString(b *encoder.BufWriter, meta *domain.Meta, key string, native string) {
	if val, ok := meta.Get(key); ok && len(val) > 0 {
		b.WriteString(",\"")
		b.Write(Escape(native))
		b.WriteString("\":\"")
		b.Write(Escape(val))
		b.WriteByte('"')
	}
}

func writeMetaList(b *encoder.BufWriter, meta *domain.Meta, key string, native string) {
	if vals, ok := meta.GetList(key); ok && len(vals) > 0 {
		b.WriteString(",\"")
		b.Write(Escape(native))
		b.WriteString("\":[\"")
		for i, val := range vals {
			if i > 0 {
				b.WriteString("\",\"")
			}
			b.Write(Escape(val))
		}
		b.WriteString("\"]")
	}
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
