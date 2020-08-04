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

// Package blob provides a parser of binary data.
package blob

import (
	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
)

func init() {
	parser.Register(&parser.Info{
		Name:         "gif",
		AltNames:     nil,
		ParseBlocks:  parseBlocks,
		ParseInlines: parseInlines,
	})
	parser.Register(&parser.Info{
		Name:         "jpeg",
		AltNames:     []string{"jpg"},
		ParseBlocks:  parseBlocks,
		ParseInlines: parseInlines,
	})
	parser.Register(&parser.Info{
		Name:         "png",
		AltNames:     nil,
		ParseBlocks:  parseBlocks,
		ParseInlines: parseInlines,
	})
}

func parseBlocks(inp *input.Input, meta *domain.Meta, syntax string) ast.BlockSlice {
	title, _ := meta.Get(domain.MetaKeyTitle)
	return ast.BlockSlice{
		&ast.BLOBNode{
			Title:  title,
			Syntax: syntax,
			Blob:   []byte(inp.Src),
		},
	}
}

func parseInlines(inp *input.Input, syntax string) ast.InlineSlice {
	return ast.InlineSlice{}
}
