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
	"context"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

// ListTagsPort is the interface used by this use case.
type ListTagsPort interface {
	// SelectMeta returns all zettel meta data that match the selection
	// criteria. The result is ordered by descending zettel id.
	SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error)
}

// ListTags is the data for this use case.
type ListTags struct {
	store ListTagsPort
}

// NewListTags creates a new use case.
func NewListTags(port ListTagsPort) ListTags {
	return ListTags{store: port}
}

// TagData associates tags with a list of all zettel meta that use this tag
type TagData map[string][]*domain.Meta

// Run executes the use case.
func (uc ListTags) Run(ctx context.Context, minCount int) (TagData, error) {
	metas, err := uc.store.SelectMeta(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	result := make(TagData)
	for _, meta := range metas {
		if tl, ok := meta.GetList(domain.MetaKeyTags); ok && len(tl) > 0 {
			for _, t := range tl {
				result[t] = append(result[t], meta)
			}
		}
	}
	if minCount > 1 {
		for t, ms := range result {
			if len(ms) < minCount {
				delete(result, t)
			}
		}
	}
	return result, nil
}
