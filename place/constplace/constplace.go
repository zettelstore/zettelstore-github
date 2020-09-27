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

// Package constplace places zettel inside the executable.
package constplace

import (
	"context"
	"errors"
	"net/url"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

func init() {
	place.Register(
		"globals",
		func(u *url.URL) (place.Place, error) {
			return &constPlace{u: u, zettel: constZettelMap}, nil
		})
}

type constHeader map[string]string

func makeMeta(zid domain.ZettelID, h constHeader) *domain.Meta {
	m := domain.NewMeta(zid)
	for k, v := range h {
		m.Set(k, v)
	}
	m.Freeze()
	return m
}

type constZettel struct {
	header  constHeader
	content domain.Content
}

type constPlace struct {
	u      *url.URL
	zettel map[domain.ZettelID]constZettel
}

// Location returns some information where the place is located.
func (cp *constPlace) Location() string {
	return cp.u.String()
}

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (cp *constPlace) Start(ctx context.Context) error {
	return nil
}

// Stop the started place. Now only the Start() function is allowed.
func (cp *constPlace) Stop(ctx context.Context) error {
	return nil
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (cp *constPlace) RegisterChangeObserver(f place.ObserverFunc) {
	// This place never changes anything. So ignore the registration.
}

func (cp *constPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	return domain.InvalidZettelID, errReadOnly
}

// GetZettel retrieves a specific zettel.
func (cp *constPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	if z, ok := cp.zettel[zid]; ok {
		return domain.Zettel{Meta: makeMeta(zid, z.header), Content: z.content}, nil
	}
	return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (cp *constPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	if z, ok := cp.zettel[zid]; ok {
		return makeMeta(zid, z.header), nil
	}
	return nil, &place.ErrUnknownID{Zid: zid}
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (cp *constPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) (res []*domain.Meta, err error) {
	hasMatch := place.CreateFilterFunc(f)
	for zid, zettel := range cp.zettel {
		meta := makeMeta(zid, zettel.header)
		if hasMatch(meta) {
			res = append(res, meta)
		}
	}
	return place.ApplySorter(res, s), nil
}

var errReadOnly = errors.New("Read-only place")

func (cp *constPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	return errReadOnly
}

// DeleteZettel removes the zettel from the place.
func (cp *constPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	return errReadOnly
}

// Rename changes the current id to a new id.
func (cp *constPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	return errReadOnly
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (cp *constPlace) Reload(ctx context.Context) error { return nil }
