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

// Package markdown provides a parser for markdown.
package markdown

import (
	"strings"
	"testing"

	"zettelstore.de/z/ast"
)

func TestSplitText(t *testing.T) {
	var testcases = []struct {
		text string
		exp  string
	}{
		{"", ""},
		{"abc", "Tabc"},
		{" ", "S "},
		{"abc def", "TabcS Tdef"},
		{"abc def ", "TabcS TdefS "},
		{" abc def ", "S TabcS TdefS "},
	}
	for i, tc := range testcases {
		var sb strings.Builder
		for _, in := range splitText(tc.text) {
			switch n := in.(type) {
			case *ast.TextNode:
				sb.WriteByte('T')
				sb.WriteString(n.Text)
			case *ast.SpaceNode:
				sb.WriteByte('S')
				sb.WriteString(n.Lexeme)
			default:
				sb.WriteByte('Q')
			}
		}
		got := sb.String()
		if tc.exp != got {
			t.Errorf("TC=%d, text=%q, exp=%q, got=%q", i, tc.text, tc.exp, got)
		}
	}
}
