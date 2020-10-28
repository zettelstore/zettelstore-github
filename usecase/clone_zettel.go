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

// CloneZettel is the data for this use case.
type CloneZettel struct{}

// NewCloneZettel creates a new use case.
func NewCloneZettel() CloneZettel {
	return CloneZettel{}
}

// Run executes the use case.
func (uc CloneZettel) Run(origZettel domain.Zettel) domain.Zettel {
	meta := origZettel.Meta.Clone()
	if title, ok := meta.Get(domain.MetaKeyTitle); ok {
		if len(title) > 0 {
			title = "Copy of " + title
		} else {
			title = "Copy"
		}
		meta.Set(domain.MetaKeyTitle, title)
	}
	return domain.Zettel{Meta: meta, Content: origZettel.Content}
}
