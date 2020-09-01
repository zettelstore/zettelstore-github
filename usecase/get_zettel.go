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
)

// GetZettelPort is the interface used by this use case.
type GetZettelPort interface {
	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)
}

// GetZettel is the data for this use case.
type GetZettel struct {
	store GetZettelPort
}

// NewGetZettel creates a new use case.
func NewGetZettel(port GetZettelPort) GetZettel {
	return GetZettel{store: port}
}

// Run executes the use case.
func (uc GetZettel) Run(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	return uc.store.GetZettel(ctx, zid)
}
