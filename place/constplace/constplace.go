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
	"net/url"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

func init() {
	place.Register(
		"globals",
		func(u *url.URL, next place.Place) (place.Place, error) {
			return &constPlace{u: u, next: next, zettel: constZettelMap}, nil
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
	next   place.Place
	zettel map[domain.ZettelID]constZettel
}

func (cp *constPlace) Next() place.Place { return cp.next }

// Location returns some information where the place is located.
func (cp *constPlace) Location() string {
	return cp.u.String()
}

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (cp *constPlace) Start(ctx context.Context) error {
	if cp.next != nil {
		return cp.next.Start(ctx)
	}
	return nil
}

// Stop the started place. Now only the Start() function is allowed.
func (cp *constPlace) Stop(ctx context.Context) error {
	if cp.next != nil {
		return cp.next.Stop(ctx)
	}
	return nil
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (cp *constPlace) RegisterChangeObserver(f place.ObserverFunc) {
	// This place never changes anything. So ignore the registration.
	if cp.next != nil {
		cp.next.RegisterChangeObserver(f)
	}
}

func (cp *constPlace) CanCreateZettel(ctx context.Context) bool { return false }

func (cp *constPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	return domain.InvalidZettelID, place.ErrReadOnly
}

// GetZettel retrieves a specific zettel.
func (cp *constPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	if z, ok := cp.zettel[zid]; ok {
		return domain.Zettel{Meta: makeMeta(zid, z.header), Content: z.content}, nil
	}
	if cp.next != nil {
		return cp.next.GetZettel(ctx, zid)
	}
	return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (cp *constPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	if z, ok := cp.zettel[zid]; ok {
		return makeMeta(zid, z.header), nil
	}
	if cp.next != nil {
		return cp.next.GetMeta(ctx, zid)
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
	if cp.next != nil {
		other, err := cp.next.SelectMeta(ctx, f, nil)
		if err != nil {
			return nil, err
		}
		return place.MergeSorted(place.ApplySorter(res, nil), other, s), nil
	}
	return place.ApplySorter(res, s), nil
}

func (cp *constPlace) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool {
	if _, ok := cp.zettel[zettel.Meta.Zid]; !ok && cp.next != nil {
		return cp.next.CanUpdateZettel(ctx, zettel)
	}
	return false
}

func (cp *constPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	if _, ok := cp.zettel[zettel.Meta.Zid]; !ok && cp.next != nil {
		return cp.next.UpdateZettel(ctx, zettel)
	}
	return place.ErrReadOnly
}

func (cp *constPlace) CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool {
	if _, ok := cp.zettel[zid]; !ok && cp.next != nil {
		return cp.next.CanDeleteZettel(ctx, zid)
	}
	return false
}

// DeleteZettel removes the zettel from the place.
func (cp *constPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	if _, ok := cp.zettel[zid]; !ok && cp.next != nil {
		return cp.next.DeleteZettel(ctx, zid)
	}
	return place.ErrReadOnly
}

func (cp *constPlace) CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool {
	if _, ok := cp.zettel[zid]; !ok {
		return cp.next == nil || cp.next.CanRenameZettel(ctx, zid)
	}
	return false
}

// Rename changes the current id to a new id.
func (cp *constPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	if _, ok := cp.zettel[curZid]; !ok {
		if cp.next != nil {
			return cp.next.RenameZettel(ctx, curZid, newZid)
		}
		return nil
	}
	return place.ErrReadOnly
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (cp *constPlace) Reload(ctx context.Context) error { return nil }
