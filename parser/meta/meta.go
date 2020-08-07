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

// Package meta provides a parser for meta data.
package meta

import (
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
)

func init() {
	parser.Register(&parser.Info{
		Name:         "meta",
		AltNames:     []string{},
		ParseBlocks:  parseBlocks,
		ParseInlines: parseInlines,
	})
}

func parseBlocks(inp *input.Input, meta *domain.Meta, syntax string) ast.BlockSlice {
	descrlist := &ast.DescriptionListNode{}
	for _, p := range meta.Pairs() {
		descrlist.Descriptions = append(descrlist.Descriptions, getDescription(p.Key, p.Value))
	}
	return ast.BlockSlice{descrlist}
}

func getDescription(key, value string) ast.Description {
	makeLink := domain.KeyType(key) == domain.MetaTypeID
	return ast.Description{
		Term: ast.InlineSlice{&ast.TextNode{Text: key}},
		Descriptions: []ast.DescriptionSlice{
			ast.DescriptionSlice{
				&ast.ParaNode{
					Inlines: convertToInlineSlice(value, makeLink),
				},
			},
		},
	}
}

func convertToInlineSlice(value string, makeLink bool) ast.InlineSlice {
	sl := strings.Fields(value)
	if len(sl) == 0 {
		return ast.InlineSlice{}
	}

	result := make(ast.InlineSlice, 0, 2*len(sl)-1)
	for i, s := range sl {
		if i > 0 {
			result = append(result, &ast.SpaceNode{Lexeme: " "})
		}
		result = append(result, &ast.TextNode{Text: s})
	}
	if makeLink {
		r := ast.ParseReference(value)
		result = ast.InlineSlice{&ast.LinkNode{Ref: r, Inlines: result}}
	}
	return result
}

func parseInlines(inp *input.Input, syntax string) ast.InlineSlice {
	inp.SkipToEOL()
	return ast.InlineSlice{
		&ast.FormatNode{
			Code:  ast.FormatSpan,
			Attrs: &ast.Attributes{Attrs: map[string]string{"class": "warning"}},
			Inlines: ast.InlineSlice{
				&ast.TextNode{Text: "parser.meta.ParseInlines:"},
				&ast.SpaceNode{Lexeme: " "},
				&ast.TextNode{Text: "not"},
				&ast.SpaceNode{Lexeme: " "},
				&ast.TextNode{Text: "possible"},
				&ast.SpaceNode{Lexeme: " "},
				&ast.TextNode{Text: "("},
				&ast.TextNode{Text: inp.Src[0:inp.Pos]},
				&ast.TextNode{Text: ")"},
			},
		},
	}
}
