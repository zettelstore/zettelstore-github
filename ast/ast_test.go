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

// Package ast_test provides the tests for the abstract syntax tree.
package ast_test

import (
	"testing"

	"zettelstore.de/z/ast"
)

func TestParseReference(t *testing.T) {
	testcases := []struct {
		link string
		err  bool
		exp  string
	}{
		{"", true, ""},
		{"123", false, "123"},
		{",://", true, ""},
	}

	for i, tc := range testcases {
		got := ast.ParseReference(tc.link)
		if got.IsValid() == tc.err {
			t.Errorf("TC=%d, expected parse error of %q: %v, but got %q", i, tc.link, tc.err, got)
		}
		if got.IsValid() && got.String() != tc.exp {
			t.Errorf("TC=%d, Reference of %q is %q, but got %q", i, tc.link, tc.exp, got)
		}
	}
}

func TestReferenceIsZettelMaterial(t *testing.T) {
	testcases := []struct {
		link       string
		isZettel   bool
		isMaterial bool
	}{
		{"", false, false},
		{"http://zettelstore.de/z/ast", false, true},
		{"12345678901234", true, false},
		{"http://12345678901234", false, true},
		{"http://zettelstore.de/z/12345678901234", false, true},
	}

	for i, tc := range testcases {
		ref := ast.ParseReference(tc.link)
		isZettel := ref.IsZettel()
		if isZettel != tc.isZettel {
			t.Errorf("TC=%d, Reference %q isZettel=%v expected, but got %v", i, tc.link, tc.isZettel, isZettel)
		}
		isMaterial := ref.IsMaterial()
		if isMaterial != tc.isMaterial {
			t.Errorf("TC=%d, Reference %q isMaterial=%v expected, but got %v", i, tc.link, tc.isMaterial, isMaterial)
		}
	}
}
