//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
	bs := Get(syntax).ParseBlocks(inp, meta, syntax)
	cleanupBlockSlice(bs)
	return bs
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
func ParseZettel(zettel domain.Zettel, syntax string) *ast.ZettelNode {
	meta := zettel.Meta
	inhMeta := config.AddDefaultValues(zettel.Meta)
	if len(syntax) == 0 {
		syntax, _ = inhMeta.Get(domain.MetaKeySyntax)
	}
	title, _ := inhMeta.Get(domain.MetaKeyTitle)
	parseMeta := inhMeta
	if syntax == "meta" {
		parseMeta = meta
	}
	return &ast.ZettelNode{
		Zettel:  zettel,
		Zid:     meta.Zid,
		InhMeta: inhMeta,
		Title:   ParseTitle(title),
		Ast:     ParseBlocks(input.NewInput(zettel.Content.AsString()), parseMeta, syntax),
	}
}
