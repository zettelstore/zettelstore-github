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

// Package policy provides some interfaces and implementation for authorizsation policies.
package policy

import (
	"context"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/web/session"
)

// pol implements a policy place.
type polPlace struct {
	place  place.Place
	policy Policy
	next   place.Place
}

// NewPlace creates a new policy place.
func NewPlace(place place.Place, policy Policy, next place.Place) place.Place {
	return &polPlace{
		place:  place,
		policy: policy,
		next:   next,
	}
}

func (pp *polPlace) Next() place.Place { return pp.next }

func (pp *polPlace) Location() string {
	return pp.place.Location()
}

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (pp *polPlace) Start(ctx context.Context) error {
	return pp.place.Start(ctx)
}

// Stop the started place. Now only the Start() function is allowed.
func (pp *polPlace) Stop(ctx context.Context) error {
	return pp.place.Stop(ctx)
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (pp *polPlace) RegisterChangeObserver(f place.ObserverFunc) {
	pp.place.RegisterChangeObserver(f)
}

func (pp *polPlace) CanCreateZettel(ctx context.Context) bool {
	return pp.place.CanCreateZettel(ctx)
}

func (pp *polPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	user := session.GetUser(ctx)
	if pp.policy.CanCreate(user, zettel.Meta) {
		return pp.place.CreateZettel(ctx, zettel)
	}
	return domain.InvalidZettelID, place.NewErrNotAuthorized("Create", user, domain.InvalidZettelID)
}

func (pp *polPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	zettel, err := pp.place.GetZettel(ctx, zid)
	if err != nil {
		return domain.Zettel{}, err
	}
	user := session.GetUser(ctx)
	if pp.policy.CanRead(user, zettel.Meta) {
		return zettel, nil
	}
	return domain.Zettel{}, place.NewErrNotAuthorized("GetZettel", user, zid)
}

// GetMeta retrieves just the meta data of a specific zettel.
func (pp *polPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	meta, err := pp.place.GetMeta(ctx, zid)
	if err != nil {
		return nil, err
	}
	user := session.GetUser(ctx)
	if pp.policy.CanRead(user, meta) {
		return meta, nil
	}
	return nil, place.NewErrNotAuthorized("GetMeta", user, zid)
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (pp *polPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error) {
	metaList, err := pp.place.SelectMeta(ctx, f, s)
	if err != nil {
		return nil, err
	}
	user := session.GetUser(ctx)
	result := make([]*domain.Meta, 0, len(metaList))
	for _, meta := range metaList {
		if pp.policy.CanRead(user, meta) {
			result = append(result, meta)
		}
	}
	return result, nil
}

func (pp *polPlace) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool {
	return pp.place.CanUpdateZettel(ctx, zettel)
}

func (pp *polPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	zid := zettel.Meta.Zid
	user := session.GetUser(ctx)
	if !zid.IsValid() {
		return &place.ErrInvalidID{Zid: zid}
	}
	// Write existing zettel
	oldMeta, err := pp.place.GetMeta(ctx, zid)
	if err != nil {
		return err
	}
	if pp.policy.CanWrite(user, oldMeta, zettel.Meta) {
		return pp.place.UpdateZettel(ctx, zettel)
	}
	return place.NewErrNotAuthorized("Write", user, zid)
}

func (pp *polPlace) CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool {
	return pp.place.CanRenameZettel(ctx, zid)
}

// Rename changes the current zid to a new zid.
func (pp *polPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	meta, err := pp.place.GetMeta(ctx, curZid)
	if err != nil {
		return err
	}
	user := session.GetUser(ctx)
	if pp.policy.CanRename(user, meta) {
		return pp.place.RenameZettel(ctx, curZid, newZid)
	}
	return place.NewErrNotAuthorized("Rename", user, curZid)
}

func (pp *polPlace) CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool {
	return pp.place.CanDeleteZettel(ctx, zid)
}

// DeleteZettel removes the zettel from the place.
func (pp *polPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	meta, err := pp.place.GetMeta(ctx, zid)
	if err != nil {
		return err
	}
	user := session.GetUser(ctx)
	if pp.policy.CanDelete(user, meta) {
		return pp.place.DeleteZettel(ctx, zid)
	}
	return place.NewErrNotAuthorized("Delete", user, zid)
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (pp *polPlace) Reload(ctx context.Context) error {
	user := session.GetUser(ctx)
	if pp.policy.CanReload(user) {
		return pp.place.Reload(ctx)
	}
	return place.NewErrNotAuthorized("Reload", user, domain.InvalidZettelID)
}
