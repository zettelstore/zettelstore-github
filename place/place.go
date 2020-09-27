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

// Package place provides a generic interface to zettel places.
package place

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"

	"zettelstore.de/z/domain"
)

// ObserverFunc is the function that will be called if something changed.
// If the first parameter, a bool, is true, then all zettel are possibly
// changed. If it has the value false, the given ZettelID will identify the
// changed zettel.
type ObserverFunc func(bool, domain.ZettelID)

// Place is implemented by all Zettel places.
type Place interface {
	// Location returns some information where the place is located.
	// Format is dependent of the place.
	Location() string

	// Start the place. Now all other functions of the place are allowed.
	// Starting an already started place is not allowed.
	Start(ctx context.Context) error

	// Stop the started place. Now only the Start() function is allowed.
	Stop(ctx context.Context) error

	// RegisterChangeObserver registers an observer that will be notified
	// if one or all zettel are found to be changed.
	RegisterChangeObserver(ObserverFunc)

	// CreateZettel creates a new zettel.
	// Returns the new zettel id (and an error indication).
	CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error)

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)

	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)

	// SelectMeta returns all zettel meta data that match the selection criteria.
	// TODO: more docs
	SelectMeta(ctx context.Context, f *Filter, s *Sorter) ([]*domain.Meta, error)

	// UpdateZettel updates an existing zettel.
	UpdateZettel(ctx context.Context, zettel domain.Zettel) error

	// DeleteZettel removes the zettel from the place.
	DeleteZettel(ctx context.Context, zid domain.ZettelID) error

	// Rename changes the current Zid to a new Zid.
	RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error

	// Reload clears all caches, reloads all internal data to reflect changes
	// that were possibly undetected.
	Reload(ctx context.Context) error
}

// ErrNotAuthorized is returned if the caller has no authorization to perform the operation.
type ErrNotAuthorized struct {
	Op   string
	User *domain.Meta
	Zid  domain.ZettelID
}

// NewErrNotAuthorized creates an new authorization error.
func NewErrNotAuthorized(op string, user *domain.Meta, zid domain.ZettelID) error {
	return &ErrNotAuthorized{
		Op:   op,
		User: user,
		Zid:  zid,
	}
}

func (err *ErrNotAuthorized) Error() string {
	if err.User == nil {
		if err.Zid.IsValid() {
			return fmt.Sprintf(
				"Operation %q on zettel %v not allowed for not authorized user",
				err.Op,
				err.Zid.Format())
		}
		return fmt.Sprintf("Operation %q not allowed for not authorized user", err.Op)
	}
	if err.Zid.IsValid() {
		return fmt.Sprintf(
			"Operation %q on zettel %v not allowed for user %v/%v",
			err.Op,
			err.Zid.Format(),
			err.User.GetDefault(domain.MetaKeyIdent, "?"),
			err.User.Zid.Format())
	}
	return fmt.Sprintf(
		"Operation %q not allowed for user %v/%v",
		err.Op,
		err.User.GetDefault(domain.MetaKeyIdent, "?"),
		err.User.Zid.Format())
}

// IsAuthError return true, if the error is of type ErrNotAuthorized.
func IsAuthError(err error) bool {
	_, ok := err.(*ErrNotAuthorized)
	return ok
}

// ErrStopped is returned if calling methods on a place that was not started.
var ErrStopped = errors.New("Place is stopped")

// ErrUnknownID is returned if the zettel id is unknown to the place.
type ErrUnknownID struct{ Zid domain.ZettelID }

func (err *ErrUnknownID) Error() string { return "Unknown Zettel id: " + err.Zid.Format() }

// ErrInvalidID is returned if the zettel id is not appropriate for the place operation.
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

// Connect returns a handle to the specified place
func Connect(rawURL string) (Place, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "dir"
	}
	if create, ok := registry[u.Scheme]; ok {
		return create(u)
	}
	return nil, &ErrInvalidScheme{u.Scheme}
}

// ErrInvalidScheme is returned if there is no place with the given scheme
type ErrInvalidScheme struct{ Scheme string }

func (err *ErrInvalidScheme) Error() string { return "Invalid scheme: " + err.Scheme }

type createFunc func(*url.URL) (Place, error)

var registry = map[string]createFunc{}

// Register the encoder for later retrieval.
func Register(scheme string, create createFunc) {
	if _, ok := registry[scheme]; ok {
		log.Fatalf("Place with scheme %q already registered", scheme)
	}
	registry[scheme] = create
}

// GetSchemes returns all registered scheme, ordered by scheme string.
func GetSchemes() []string {
	result := make([]string, 0, len(registry))
	for scheme := range registry {
		result = append(result, scheme)
	}
	sort.Strings(result)
	return result
}
