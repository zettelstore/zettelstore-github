//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
	if p := parser.Get(syntax); p != nil {
		syntax = p.Name
	}
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
