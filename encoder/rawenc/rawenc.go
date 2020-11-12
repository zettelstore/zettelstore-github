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

// Package rawenc encodes the abstract syntax tree as raw content.
package rawenc

import (
	"io"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
)

func init() {
	encoder.Register("raw", encoder.Info{
		Create: func() encoder.Encoder { return &rawEncoder{} },
	})
}

type rawEncoder struct{}

// SetOption sets an option for the encoder
func (re *rawEncoder) SetOption(option encoder.Option) {}

// WriteZettel writes the encoded zettel to the writer.
func (re *rawEncoder) WriteZettel(w io.Writer, zn *ast.ZettelNode, inhMeta bool) (int, error) {
	b := encoder.NewBufWriter(w)
	if inhMeta {
		zn.InhMeta.Write(&b)
	} else {
		zn.Zettel.Meta.Write(&b)
	}
	b.WriteByte('\n')
	b.WriteString(zn.Zettel.Content.AsString())
	length, err := b.Flush()
	return length, err
}

// WriteMeta encodes meta data as HTML5.
func (re *rawEncoder) WriteMeta(w io.Writer, meta *domain.Meta) (int, error) {
	b := encoder.NewBufWriter(w)
	meta.Write(&b)
	length, err := b.Flush()
	return length, err
}

func (re *rawEncoder) WriteContent(w io.Writer, zn *ast.ZettelNode) (int, error) {
	b := encoder.NewBufWriter(w)
	b.WriteString(zn.Zettel.Content.AsString())
	length, err := b.Flush()
	return length, err
}

// WriteBlocks writes a block slice to the writer
func (re *rawEncoder) WriteBlocks(w io.Writer, bs ast.BlockSlice) (int, error) {
	return 0, encoder.ErrNoWriteBlocks
}

// WriteInlines writes an inline slice to the writer
func (re *rawEncoder) WriteInlines(w io.Writer, is ast.InlineSlice) (int, error) {
	return 0, encoder.ErrNoWriteInlines
}
