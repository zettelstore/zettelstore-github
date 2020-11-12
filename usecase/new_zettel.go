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

// Package usecase provides (business) use cases for the zettelstore.
package usecase

import (
	"zettelstore.de/z/domain"
)

// NewZettel is the data for this use case.
type NewZettel struct{}

// NewNewZettel creates a new use case.
func NewNewZettel() NewZettel {
	return NewZettel{}
}

// Run executes the use case.
func (uc NewZettel) Run(origZettel domain.Zettel) domain.Zettel {
	meta := origZettel.Meta.Clone()
	if role, ok := meta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleNewTemplate {
		const prefix = "new-"
		for _, pair := range meta.PairsRest() {
			if key := pair.Key; len(key) > len(prefix) && key[0:len(prefix)] == prefix {
				meta.Set(key[len(prefix):], pair.Value)
				meta.Delete(key)
			}
		}
	}
	return domain.Zettel{Meta: meta, Content: origZettel.Content}
}
