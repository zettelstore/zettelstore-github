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

// Package zettelmark provides a parser for zettelmarkup.
package zettelmark

import (
	"zettelstore.de/z/ast"
)

// Internal nodes for parsing zettelmark. These will be removed in
// post-processing.

// nullItemNode specifies a removable placeholder for an item block.
type nullItemNode struct {
	ast.ItemNode
}

func (nn *nullItemNode) blockNode() {}
func (nn *nullItemNode) itemNode()  {}

// Accept a visitor and visit the node.
func (nn *nullItemNode) Accept(v ast.Visitor) {}

// nullDescriptionNode specifies a removable placeholder.
type nullDescriptionNode struct {
	ast.DescriptionNode
}

func (nn *nullDescriptionNode) blockNode()       {}
func (nn *nullDescriptionNode) descriptionNode() {}

// Accept a visitor and visit the node.
func (nn *nullDescriptionNode) Accept(v ast.Visitor) {}
