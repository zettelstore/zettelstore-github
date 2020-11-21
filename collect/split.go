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

// Package collect provides functions to collect items from a syntax tree.
package collect

import (
	"zettelstore.de/z/ast"
)

// DivideReferences divides the given list of rederences into zettel, local, and external References.
func DivideReferences(all []*ast.Reference, duplicates bool) (zettel, local, external []*ast.Reference) {
	if len(all) == 0 {
		return nil, nil, nil
	}

	mapZettel := make(map[string]bool)
	mapLocal := make(map[string]bool)
	mapExternal := make(map[string]bool)
	for _, ref := range all {
		s := ref.String()
		if ref.IsZettel() {
			if duplicates {
				zettel = append(zettel, ref)
			} else {
				if _, ok := mapZettel[s]; !ok {
					zettel = append(zettel, ref)
					mapZettel[s] = true
				}
			}
		} else if ref.IsExternal() {
			if duplicates {
				external = append(external, ref)
			} else {
				if _, ok := mapExternal[s]; !ok {
					external = append(external, ref)
					mapExternal[s] = true
				}
			}
		} else {
			if duplicates {
				local = append(local, ref)
			} else {
				if _, ok := mapLocal[s]; !ok {
					local = append(local, ref)
					mapLocal[s] = true
				}
			}
		}
	}
	return zettel, local, external
}
