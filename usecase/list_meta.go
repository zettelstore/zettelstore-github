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

// ListMetaPort is the interface used by this use case.
type ListMetaPort interface {
	// SelectMeta returns all zettel meta data that match the selection
	// criteria. The result is ordered by descending zettel id.
	SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error)
}

// ListMeta is the data for this use case.
type ListMeta struct {
	store ListMetaPort
}

// NewListMeta creates a new use case.
func NewListMeta(port ListMetaPort) ListMeta {
	return ListMeta{store: port}
}

// Run executes the use case.
func (uc ListMeta) Run(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error) {
	return uc.store.SelectMeta(ctx, f, s)
}
