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
	"fmt"

	"zettelstore.de/z/domain"
)

// ObserverFunc is the function that will be called if something changed.
// If the first parameter, a bool, is true, then all zettel are possibly
// changed. If it has the value false, the given ZettelID will identify the
// changed zettel.
type ObserverFunc func(bool, domain.ZettelID)

// Store is implemented by all Zettel stores.
type Store interface {
	// Location returns some information where the store is located.
	// Format is dependent of the store.
	Location() string

	// Start the store. Now all other functions of the store are allowed.
	// Starting an already started store is not allowed.
	Start(ctx context.Context) error

	// Stop the started store. Now only the Start() function is allowed.
	Stop(ctx context.Context) error

	// RegisterChangeObserver registers an observer that will be notified
	// if one or all zettel are found to be changed.
	RegisterChangeObserver(ObserverFunc)

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)

	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)

	// SelectMeta returns all zettel meta data that match the selection criteria.
	// TODO: more docs
	SelectMeta(ctx context.Context, f *Filter, s *Sorter) ([]*domain.Meta, error)

	// SetZettel updates an existing zettel or creates a new one.
	// It the zettel contains a valid Zid, an update operation is assumed,
	// otherwise the store must assign a new Zid for the zettel. In this case, the
	// meta data of the zettel will contain the updated Zid. The caller is
	// potentially allowed to assign an Zid itself, but at own risk.
	SetZettel(ctx context.Context, zettel domain.Zettel) error

	// DeleteZettel removes the zettel from the store.
	DeleteZettel(ctx context.Context, zid domain.ZettelID) error

	// Rename changes the current Zid to a new Zid.
	RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error

	// Reload clears all caches, reloads all internal data to reflect changes
	// that were possibly undetected.
	Reload(ctx context.Context) error
}

// ErrNotAuthorized is returned if the caller has no authorization to perform the operation.
type ErrNotAuthorized struct {
	op   string
	user *domain.Meta
	zid  domain.ZettelID
}

// NewErrNotAuthorized creates an new authorization error.
func NewErrNotAuthorized(op string, user *domain.Meta, zid domain.ZettelID) error {
	return &ErrNotAuthorized{
		op:   op,
		user: user,
		zid:  zid,
	}
}

func (err *ErrNotAuthorized) Error() string {
	if err.user == nil {
		if err.zid.IsValid() {
			return fmt.Sprintf(
				"Operation %q on zettel %v not allowed for not authorized user",
				err.op,
				err.zid.Format())
		}
		return fmt.Sprintf("Operation %q not allowed for not authorized user", err.op)
	}
	if err.zid.IsValid() {
		return fmt.Sprintf(
			"Operation %q on zettel %v not allowed for user %v/%v",
			err.op,
			err.zid.Format(),
			err.user.GetDefault(domain.MetaKeyIdent, "?"),
			err.user.Zid.Format())
	}
	return fmt.Sprintf(
		"Operation %q not allowed for user %v/%v",
		err.op,
		err.user.GetDefault(domain.MetaKeyIdent, "?"),
		err.user.Zid.Format())
}

// IsAuthError return true, if the error is of type ErrNotAuthorized.
func IsAuthError(err error) bool {
	_, ok := err.(*ErrNotAuthorized)
	return ok
}

// ErrStopped is returned if calling methods on a store that was not started.
var ErrStopped = errors.New("Store is stopped")

// ErrUnknownID is returned if the zettel id is unknown to the store.
type ErrUnknownID struct{ Zid domain.ZettelID }

func (err *ErrUnknownID) Error() string { return "Unknown Zettel id: " + err.Zid.Format() }

// ErrInvalidID is returned if the zettel id is not appropriate for the store operation.
type ErrInvalidID struct{ Zid domain.ZettelID }

func (err *ErrInvalidID) Error() string { return "Invalid Zettel id: " + err.Zid.Format() }

// Filter specifies a mechanism for selecting zettel.
type Filter struct {
	Expr   FilterExpr
	Negate bool
}

// FilterExpr is the encoding of a search filter.
type FilterExpr map[string][]string // map of keys to or-ed values

// Sorter specifies ordering and limiting a sequnce of meta data.
type Sorter struct {
	Order      string // Name of meta key. None given: use "id"
	Descending bool   // Sort by order, but descending
	Offset     int    // <= 0: no offset
	Limit      int    // <= 0: no limit
}
