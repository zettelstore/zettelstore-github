//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
