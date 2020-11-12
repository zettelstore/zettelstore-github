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

// Package ast provides the abstract syntax tree.
package ast

import (
	"net/url"

	"zettelstore.de/z/domain"
)

// ZettelNode is the root node of the abstract syntax tree.
// It is *not* part of the visitor pattern.
type ZettelNode struct {
	Zettel  domain.Zettel
	Zid     domain.ZettelID // Zettel identification.
	InhMeta *domain.Meta    // Meta data of the zettel, with inherited values.
	Title   InlineSlice     // Zettel title is a sequence of inline nodes.
	Ast     BlockSlice      // Zettel abstract syntax tree is a sequence of block nodes.
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
		if _, err := domain.ParseZettelID(u.Path); err == nil {
			return &Reference{URL: u, Value: s, State: RefStateZettel}
		}
		if len(u.Path) > 1 && u.Path[0] == '/' {
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
	case RefStateZettel, RefStateZettelFound, RefStateZettelBroken:
		return true
	}
	return false
}

// IsLocal returns true if reference is local
func (r *Reference) IsLocal() bool { return r.State == RefStateLocal }

// IsExternal returns true if it is a referencen to external material.
func (r *Reference) IsExternal() bool { return r.State == RefStateExternal }
