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

// Package parser provides a generic interface to a range of different parsers.
package parser

import (
	"log"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
)

// Info describes a single parser.
//
// Before ParseBlocks() or ParseInlines() is called, ensure the input stream to
// be valid. This can ce achieved on calling inp.Next() after the input stream
// was created.
type Info struct {
	Name         string
	AltNames     []string
	ParseBlocks  func(*input.Input, *domain.Meta, string) ast.BlockSlice
	ParseInlines func(*input.Input, string) ast.InlineSlice
}

var registry = map[string]*Info{}

// Register the parser (info) for later retrieval.
func Register(pi *Info) *Info {
	if _, ok := registry[pi.Name]; ok {
		log.Fatalf("Parser %q already registered", pi.Name)
	}
	registry[pi.Name] = pi
	for _, alt := range pi.AltNames {
		if _, ok := registry[alt]; ok {
			log.Fatalf("Parser %q already registered", alt)
		}
		registry[alt] = pi
	}
	return pi
}

// Get the parser (info) by name. If name not found, use a default parser.
func Get(name string) *Info {
	if pi := registry[name]; pi != nil {
		return pi
	}
	if pi := registry["plain"]; pi != nil {
		return pi
	}
	log.Printf("No parser for %q found", name)
	panic("No default parser registered")
}

// ParseBlocks parses some input and returns a slice of block nodes.
func ParseBlocks(inp *input.Input, meta *domain.Meta, syntax string) ast.BlockSlice {
	return Get(syntax).ParseBlocks(inp, meta, syntax)
}

// ParseInlines parses some input and returns a slice of inline nodes.
func ParseInlines(inp *input.Input, syntax string) ast.InlineSlice {
	return Get(syntax).ParseInlines(inp, syntax)
}

// ParseTitle parses the title of a zettel, always as Zettelmarkup
func ParseTitle(title string) ast.InlineSlice {
	return ParseInlines(input.NewInput(title), "zmk")
}

// ParseZettel parses the zettel based on the syntax.
func ParseZettel(zettel domain.Zettel, syntax string) (*ast.Zettel, *domain.Meta) {
	meta := config.Config.AddDefaultValues(zettel.Meta)
	if len(syntax) == 0 {
		syntax, _ = meta.Get(domain.MetaKeySyntax)
	}
	title, _ := meta.Get(domain.MetaKeyTitle)
	id := meta.ID
	z := &ast.Zettel{
		ID:      id,
		Meta:    zettel.Meta,
		Content: zettel.Content,
		Title:   ParseTitle(title),
		Ast:     ParseBlocks(input.NewInput(zettel.Content.AsString()), zettel.Meta, syntax),
	}
	return z, meta
}
