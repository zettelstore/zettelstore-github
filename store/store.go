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

// Package store provides a generic interface to zettel stores.
package store

import (
	"context"
	"errors"

	"zettelstore.de/z/domain"
)

// Store is implemented by all Zettel stores.
type Store interface {
	// SetParentStore is called when the store is part of a bigger store.
	SetParentStore(parent Store)

	// Location returns some information where the store is located.
	// Format is dependent of the store.
	Location() string

	// Start the store. Now all other functions of the store are allowed.
	// Starting an already started store is not allowed.
	Start(ctx context.Context) error

	// Stop the started store. Now only the Start() function is allowed.
	Stop(ctx context.Context) error

	// RegisterChangeObserver registers an observer that will be notified
	// if a zettel was found to be changed. If the id is empty, all zettel are
	// possibly changed.
	RegisterChangeObserver(func(domain.ZettelID))

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, id domain.ZettelID) (domain.Zettel, error)

	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, id domain.ZettelID) (*domain.Meta, error)

	// SelectMeta returns all zettel meta data that match the selection criteria.
	// TODO: more docs
	SelectMeta(ctx context.Context, f *Filter, s *Sorter) ([]*domain.Meta, error)

	// SetZettel updates an existing zettel or creates a new one.
	// It the zettel contains a valid ID, an update operation is assumed,
	// otherwise the store must assign a new ID for the zettel. In this case, the
	// meta data of the zettel will contain the updated ID. The caller is
	// potentially allowed to assign an ID itself, but at own risk.
	SetZettel(ctx context.Context, zettel domain.Zettel) error

	// DeleteZettel removes the zettel from the store.
	DeleteZettel(ctx context.Context, id domain.ZettelID) error

	// Rename changes the current ID to a new ID.
	RenameZettel(ctx context.Context, curID, newID domain.ZettelID) error

	// Reload clears all caches, reloads all internal data to reflect changes
	// that were possibly undetected.
	Reload(ctx context.Context) error
}

// ErrStopped is returned if calling methods on a store that was not started.
var ErrStopped = errors.New("Store is stopped")

// ErrUnknownID is returned if the zettel ID is unknown to the store.
type ErrUnknownID struct{ ID domain.ZettelID }

func (err *ErrUnknownID) Error() string { return "Unknown Zettel ID: " + string(err.ID) }

// ErrInvalidID is returned if the zettel ID is not appropriate for the store operation.
type ErrInvalidID struct{ ID domain.ZettelID }

func (err *ErrInvalidID) Error() string { return "Invalid Zettel ID: " + string(err.ID) }

// Filter specifies a mechanism for selecting zettel.
type Filter struct {
	Expr   FilterExpr
	Negate bool
}

// FilterExpr is the encoding of a search filter.
type FilterExpr map[string][]string // map of keys to or-ed values

// Sorter specifies ordering and limiting a sequnce of meta data.
type Sorter struct {
	Order  string // Name of meta key. None given: use "id". If key starts with "-" use descending order.
	Offset int    // <= 0: no offset
	Limit  int    // <= 0: no limit
}
