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
	"sync"

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

// Parser is the generic part of the parser system.
type Parser struct {
	muBlocks    sync.RWMutex
	blockCache  map[string]ast.BlockSlice
	muInlines   sync.RWMutex
	inlineCache map[string]ast.InlineSlice
}

// New creates a new parser system.
func New() *Parser {
	p := &Parser{}
	return p
}

type parserStore interface {
	// RegisterChangeObserver registers an observer that will be notified
	// if a zettel was found to be changed. If the id is empty, all zettel are
	// possibly changed.
	RegisterChangeObserver(func(domain.ZettelID))
}

// InitCache allows parse results to be cached.
func (p *Parser) InitCache(s parserStore) {
	//TODO: add cache strategy, e.g. max content size, num entries, duration, ...
	p.observe("")
	s.RegisterChangeObserver(p.observe)
}

func (p *Parser) observe(id domain.ZettelID) {
	// Remove everything, regardless of id.
	p.muBlocks.Lock()
	p.blockCache = make(map[string]ast.BlockSlice, len(p.blockCache))
	p.muBlocks.Unlock()
	p.muInlines.Lock()
	p.inlineCache = make(map[string]ast.InlineSlice, len(p.inlineCache))
	p.muInlines.Unlock()
}

// ParseBlocks parses some input and returns a slice of block nodes.
func (p *Parser) ParseBlocks(id domain.ZettelID, inp *input.Input, meta *domain.Meta, syntax string) ast.BlockSlice {
	key := string(id) + syntax
	p.muBlocks.RLock()
	if p.blockCache != nil {
		bs, ok := p.blockCache[key]
		if ok {
			p.muBlocks.RUnlock()
			return bs
		}
	}
	bs := Get(syntax).ParseBlocks(inp, meta, syntax)
	if p.blockCache != nil {
		p.muBlocks.RUnlock()
		p.muBlocks.Lock()
		if len(id) > 0 {
			p.blockCache[key] = bs
		}
		p.muBlocks.Unlock()
	} else {
		p.muBlocks.RUnlock()
	}
	return bs
}

// ParseInlines parses some input and returns a slice of inline nodes.
func (p *Parser) ParseInlines(id domain.ZettelID, inp *input.Input, syntax string) ast.InlineSlice {
	key := string(id) + syntax
	p.muInlines.RLock()
	if p.inlineCache != nil {
		is, ok := p.inlineCache[key]
		if ok {
			p.muInlines.RUnlock()
			return is
		}
	}
	is := Get(syntax).ParseInlines(inp, syntax)
	if p.inlineCache != nil {
		p.muInlines.RUnlock()
		p.muInlines.Lock()
		if len(id) > 0 {
			p.inlineCache[key] = is
		}
		p.muInlines.Unlock()
	} else {
		p.muInlines.RUnlock()
	}
	return is
}

// ParseTitle parses the title of a zettel, always as Zettelmarkup
func (p *Parser) ParseTitle(id domain.ZettelID, inp *input.Input) ast.InlineSlice {
	return p.ParseInlines(id, inp, "zmk")
}

// ParseZettel parses the zettel based on the syntax.
func (p *Parser) ParseZettel(zettel domain.Zettel, syntax string) (*ast.Zettel, *domain.Meta) {
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
		Title:   p.ParseTitle(id, input.NewInput(title)),
		Ast:     p.ParseBlocks(id, input.NewInput(zettel.Content.AsString()), zettel.Meta, syntax),
	}
	return z, meta
}
