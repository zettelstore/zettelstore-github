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

// RenameZettelPort is the interface used by this use case.
type RenameZettelPort interface {
	// Rename changes the current id to a new id.
	RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error
}

// RenameZettel is the data for this use case.
type RenameZettel struct {
	store RenameZettelPort
}

// NewRenameZettel creates a new use case.
func NewRenameZettel(port RenameZettelPort) RenameZettel {
	return RenameZettel{store: port}
}

// Run executes the use case.
func (uc RenameZettel) Run(ctx context.Context, curID, newID domain.ZettelID) error {
	return uc.store.RenameZettel(ctx, curID, newID)
}
