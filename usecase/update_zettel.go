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

// UpdateZettelPort is the interface used by this use case.
type UpdateZettelPort interface {
	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)

	// SetZettel updates an existing zettel or creates a new one.
	SetZettel(ctx context.Context, zettel domain.Zettel) error
}

// UpdateZettel is the data for this use case.
type UpdateZettel struct {
	store UpdateZettelPort
}

// NewUpdateZettel creates a new use case.
func NewUpdateZettel(port UpdateZettelPort) UpdateZettel {
	return UpdateZettel{store: port}
}

// Run executes the use case.
func (uc UpdateZettel) Run(ctx context.Context, zettel domain.Zettel) error {
	meta := zettel.Meta
	oldZettel, err := uc.store.GetZettel(ctx, meta.Zid)
	if err != nil {
		return err
	}
	if zettel.Equal(oldZettel) {
		return nil
	}
	meta.YamlSep = oldZettel.Meta.YamlSep
	if meta.Zid == domain.ConfigurationID {
		meta.Set(domain.MetaKeySyntax, "meta")
	}
	return uc.store.SetZettel(ctx, zettel)
}
