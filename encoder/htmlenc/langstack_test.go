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

// Package htmlenc encodes the abstract syntax tree into HTML5.
package htmlenc

import (
	"testing"

	"zettelstore.de/z/ast"
)

func TestStackSimple(t *testing.T) {
	exp := "de"
	s := newLangStack(exp)
	if got := s.top(); got != exp {
		t.Errorf("Init: expected %q, but got %q", exp, got)
		return
	}

	a := &ast.Attributes{}
	s.push(a)
	if got := s.top(); exp != got {
		t.Errorf("Empty push: expected %q, but got %q", exp, got)
	}

	exp2 := "en"
	a = a.Set("lang", exp2)
	s.push(a)
	if got := s.top(); exp2 != got {
		t.Errorf("Full push: expected %q, but got %q", exp2, got)
	}

	s.pop()
	if got := s.top(); exp != got {
		t.Errorf("pop: expected %q, but got %q", exp, got)
	}
}
