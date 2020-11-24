//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
		isExternal bool
	}{
		{"", false, false},
		{"http://zettelstore.de/z/ast", false, true},
		{"12345678901234", true, false},
		{"12345678901234#local", true, false},
		{"http://12345678901234", false, true},
		{"http://zettelstore.de/z/12345678901234", false, true},
		{"http://zettelstore.de/12345678901234", false, true},
		{"/12345678901234", false, false},
	}

	for i, tc := range testcases {
		ref := ast.ParseReference(tc.link)
		isZettel := ref.IsZettel()
		if isZettel != tc.isZettel {
			t.Errorf("TC=%d, Reference %q isZettel=%v expected, but got %v", i, tc.link, tc.isZettel, isZettel)
		}
		isExternal := ref.IsExternal()
		if isExternal != tc.isExternal {
			t.Errorf("TC=%d, Reference %q isExternal=%v expected, but got %v", i, tc.link, tc.isExternal, isExternal)
		}
	}
}
