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

// Package policystore provides a store that checks for authorization policies.
package policystore

import (
	"context"

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/web/session"
)

// pol implements a policy store.
type polStore struct {
	store  store.Store
	policy policy.Policy
}

// NewStore creates a new policy store.
func NewStore(store store.Store, policy policy.Policy) store.Store {
	return &polStore{
		store:  store,
		policy: policy,
	}
}

func (ps *polStore) Location() string {
	return ps.store.Location()
}

// Start the store. Now all other functions of the store are allowed.
// Starting an already started store is not allowed.
func (ps *polStore) Start(ctx context.Context) error {
	return ps.store.Start(ctx)
}

// Stop the started store. Now only the Start() function is allowed.
func (ps *polStore) Stop(ctx context.Context) error {
	return ps.store.Stop(ctx)
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (ps *polStore) RegisterChangeObserver(f store.ObserverFunc) {
	ps.store.RegisterChangeObserver(f)
}

// GetZettel reads the zettel from a file.
func (ps *polStore) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	zettel, err := ps.store.GetZettel(ctx, zid)
	if err != nil {
		return domain.Zettel{}, err
	}
	if ps.policy.CanRead(session.GetUser(ctx), zettel.Meta) {
		return zettel, nil
	}
	return domain.Zettel{}, store.ErrNotAuthorized
}

// GetMeta retrieves just the meta data of a specific zettel.
func (ps *polStore) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	meta, err := ps.store.GetMeta(ctx, zid)
	if err != nil {
		return nil, err
	}
	if ps.policy.CanRead(session.GetUser(ctx), meta) {
		return meta, nil
	}
	return nil, store.ErrNotAuthorized
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (ps *polStore) SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error) {
	metaList, err := ps.store.SelectMeta(ctx, f, s)
	if err != nil {
		return nil, err
	}
	user := session.GetUser(ctx)
	result := make([]*domain.Meta, 0, len(metaList))
	for _, meta := range metaList {
		if ps.policy.CanRead(user, meta) {
			result = append(result, meta)
		}
	}
	return result, nil
}

// SetZettel stores new data for a zettel.
func (ps *polStore) SetZettel(ctx context.Context, zettel domain.Zettel) error {
	zid := zettel.Meta.Zid
	if zid.IsValid() {
		// Write existing zettel
		oldMeta, err := ps.store.GetMeta(ctx, zid)
		if err != nil {
			return err
		}
		if ps.policy.CanWrite(session.GetUser(ctx), oldMeta, zettel.Meta) {
			return ps.store.SetZettel(ctx, zettel)
		}
	} else {
		// Create new zettel
		if ps.policy.CanCreate(session.GetUser(ctx), zettel.Meta) {
			return ps.store.SetZettel(ctx, zettel)
		}
	}
	return store.ErrNotAuthorized
}

// Rename changes the current zid to a new zid.
func (ps *polStore) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	meta, err := ps.store.GetMeta(ctx, curZid)
	if err != nil {
		return err
	}
	if ps.policy.CanRename(session.GetUser(ctx), meta) {
		return ps.store.RenameZettel(ctx, curZid, newZid)
	}
	return store.ErrNotAuthorized
}

// DeleteZettel removes the zettel from the store.
func (ps *polStore) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	meta, err := ps.store.GetMeta(ctx, zid)
	if err != nil {
		return err
	}
	if ps.policy.CanDelete(session.GetUser(ctx), meta) {
		return ps.store.DeleteZettel(ctx, zid)
	}
	return store.ErrNotAuthorized
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (ps *polStore) Reload(ctx context.Context) error {
	if ps.policy.CanReload(session.GetUser(ctx)) {
		return ps.store.Reload(ctx)
	}
	return store.ErrNotAuthorized
}
