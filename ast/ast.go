//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package ast provides the abstract syntax tree.
package ast

import (
	"net/url"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
)

// ZettelNode is the root node of the abstract syntax tree.
// It is *not* part of the visitor pattern.
type ZettelNode struct {
	Zettel  domain.Zettel
	Zid     id.ZettelID // Zettel identification.
	InhMeta *meta.Meta  // Meta data of the zettel, with inherited values.
	Title   InlineSlice // Zettel title is a sequence of inline nodes.
	Ast     BlockSlice  // Zettel abstract syntax tree is a sequence of block nodes.
}

// Node is the interface, all nodes must implement.
type Node interface {
	Accept(v Visitor)
}

// BlockNode is the interface that all block nodes must implement.
type BlockNode interface {
	Node
	blockNode()
}

// BlockSlice is a slice of BlockNodes.
type BlockSlice []BlockNode

// ItemNode is a node that can occur as a list item.
type ItemNode interface {
	BlockNode
	itemNode()
}

// ItemSlice is a slice of ItemNodes.
type ItemSlice []ItemNode

// DescriptionNode is a node that contains just textual description.
type DescriptionNode interface {
	ItemNode
	descriptionNode()
}

// DescriptionSlice is a slice of DescriptionNodes.
type DescriptionSlice []DescriptionNode

// InlineNode is the interface that all inline nodes must implement.
type InlineNode interface {
	Node
	inlineNode()
}

// InlineSlice is a slice of InlineNodes.
type InlineSlice []InlineNode

// Reference is a reference to external or internal material.
type Reference struct {
	URL   *url.URL
	Value string
	State RefState
}

// RefState indicates the state of the reference.
type RefState int

// Constants for RefState
const (
	RefStateInvalid      RefState = iota // Invalid URL
	RefStateZettel                       // Valid reference to an internal zettel
	RefStateZettelSelf                   // Valid reference to same zettel with a fragment
	RefStateZettelFound                  // Valid reference to an existing internal zettel
	RefStateZettelBroken                 // Valid reference to a non-existing internal zettel
	RefStateLocal                        // Valid reference to a non-zettel, but local hosted
	RefStateExternal                     // Valid reference to external material
)

// ParseReference parses a string and returns a reference.
func ParseReference(s string) *Reference {
	if len(s) == 0 {
		return &Reference{URL: nil, Value: s, State: RefStateInvalid}
	}
	u, err := url.Parse(s)
	if err != nil {
		return &Reference{URL: nil, Value: s, State: RefStateInvalid}
	}
	if len(u.Scheme)+len(u.Opaque)+len(u.Host) == 0 && u.User == nil {
		if _, err := id.ParseZettelID(u.Path); err == nil {
			return &Reference{URL: u, Value: s, State: RefStateZettel}
		}
		if u.Path == "" && u.Fragment != "" {
			return &Reference{URL: u, Value: s, State: RefStateZettelSelf}
		}
		if u.Path != "" && u.Path[0] == '/' {
			return &Reference{URL: u, Value: s, State: RefStateLocal}
		}
	}
	return &Reference{URL: u, Value: s, State: RefStateExternal}
}

// String returns the string representation of a reference.
func (r Reference) String() string {
	if r.URL != nil {
		return r.URL.String()
	}
	return r.Value
}

// IsValid returns true if reference is valid
func (r *Reference) IsValid() bool { return r.State != RefStateInvalid }

// IsZettel returns true if it is a referencen to a local zettel.
func (r *Reference) IsZettel() bool {
	switch r.State {
	case RefStateZettel, RefStateZettelSelf, RefStateZettelFound, RefStateZettelBroken:
		return true
	}
	return false
}

// IsLocal returns true if reference is local
func (r *Reference) IsLocal() bool { return r.State == RefStateLocal }

// IsExternal returns true if it is a referencen to external material.
func (r *Reference) IsExternal() bool { return r.State == RefStateExternal }
