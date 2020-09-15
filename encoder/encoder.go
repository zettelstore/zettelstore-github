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
	"errors"
	"io"
	"log"
	"sort"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
)

// Encoder is an interface that allows to encode different parts of a zettel.
type Encoder interface {
	SetOption(Option)

	WriteZettel(io.Writer, *ast.Zettel) (int, error)
	WriteMeta(io.Writer, *domain.Meta) (int, error)
	WriteContent(io.Writer, *ast.Zettel) (int, error)
	WriteBlocks(io.Writer, ast.BlockSlice) (int, error)
	WriteInlines(io.Writer, ast.InlineSlice) (int, error)
}

// Some errors to signal when WriteBlocks/WriteInlines are not implemented.
var (
	ErrNoWriteBlocks  = errors.New("Method WriteBlocks is not implemented")
	ErrNoWriteInlines = errors.New("Method WriteInlines is not implemented")
)

// Option allows to configure an encoder
type Option interface {
	Name() string
}

// Create builds a new encoder with the given options.
func Create(format string, options ...Option) Encoder {
	if create, ok := registry[format]; ok {
		enc := create()
		for _, opt := range options {
			enc.SetOption(opt)
		}
		return enc
	}
	return nil
}

type createFunc func() Encoder

var registry = map[string]createFunc{}

// Register the encoder for later retrieval.
func Register(format string, create createFunc) {
	if _, ok := registry[format]; ok {
		log.Fatalf("Writer with format %q already registered", format)
	}
	registry[format] = create
}

// GetFormats returns all registered formats, ordered by format name.
func GetFormats() []string {
	result := make([]string, 0, len(registry))
	for format := range registry {
		result = append(result, format)
	}
	sort.Strings(result)
	return result
}
