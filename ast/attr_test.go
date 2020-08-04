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
package ast_test

import (
	"testing"

	"zettelstore.de/z/ast"
)

func TestHasDefault(t *testing.T) {
	attr := &ast.Attributes{}
	if attr.HasDefault() {
		t.Error("Should not have default attr")
	}
	attr = &ast.Attributes{Attrs: map[string]string{"-": "value"}}
	if !attr.HasDefault() {
		t.Error("Should have default attr")
	}
}

func TestAttrClone(t *testing.T) {
	orig := &ast.Attributes{}
	clone := orig.Clone()
	if len(clone.Attrs) > 0 {
		t.Error("Attrs must be empty")
	}

	orig = &ast.Attributes{Attrs: map[string]string{"": "0", "-": "1", "a": "b"}}
	clone = orig.Clone()
	m := clone.Attrs
	if m[""] != "0" || m["-"] != "1" || m["a"] != "b" || len(m) != len(orig.Attrs) {
		t.Error("Wrong cloned map")
	}
	m["a"] = "c"
	if orig.Attrs["a"] != "b" {
		t.Error("Aliased map")
	}
}
