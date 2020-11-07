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

import "zettelstore.de/z/ast"

type langStack struct {
	items []string
}

func newLangStack(lang string) langStack {
	items := make([]string, 1, 16)
	items[0] = lang
	return langStack{items}
}

func (s langStack) top() string { return s.items[len(s.items)-1] }

func (s *langStack) pop() { s.items = s.items[0 : len(s.items)-1] }

func (s *langStack) push(attrs *ast.Attributes) {
	if value, ok := attrs.Get("lang"); ok {
		s.items = append(s.items, value)
	} else {
		s.items = append(s.items, s.top())
	}
}
