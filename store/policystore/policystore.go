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

func (ps *polStore) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	user := session.GetUser(ctx)
	if ps.policy.CanCreate(user, zettel.Meta) {
		return ps.store.CreateZettel(ctx, zettel)
	}
	return domain.InvalidZettelID, store.NewErrNotAuthorized("Create", user, domain.InvalidZettelID)
}

func (ps *polStore) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	zettel, err := ps.store.GetZettel(ctx, zid)
	if err != nil {
		return domain.Zettel{}, err
	}
	user := session.GetUser(ctx)
	if ps.policy.CanRead(user, zettel.Meta) {
		return zettel, nil
	}
	return domain.Zettel{}, store.NewErrNotAuthorized("GetZettel", user, zid)
}

// GetMeta retrieves just the meta data of a specific zettel.
func (ps *polStore) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	meta, err := ps.store.GetMeta(ctx, zid)
	if err != nil {
		return nil, err
	}
	user := session.GetUser(ctx)
	if ps.policy.CanRead(user, meta) {
		return meta, nil
	}
	return nil, store.NewErrNotAuthorized("GetMeta", user, zid)
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

func (ps *polStore) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	zid := zettel.Meta.Zid
	user := session.GetUser(ctx)
	if !zid.IsValid() {
		return &store.ErrInvalidID{Zid: zid}
	}
	// Write existing zettel
	oldMeta, err := ps.store.GetMeta(ctx, zid)
	if err != nil {
		return err
	}
	if ps.policy.CanWrite(user, oldMeta, zettel.Meta) {
		return ps.store.UpdateZettel(ctx, zettel)
	}
	return store.NewErrNotAuthorized("Write", user, zid)
}

// Rename changes the current zid to a new zid.
func (ps *polStore) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	meta, err := ps.store.GetMeta(ctx, curZid)
	if err != nil {
		return err
	}
	user := session.GetUser(ctx)
	if ps.policy.CanRename(user, meta) {
		return ps.store.RenameZettel(ctx, curZid, newZid)
	}
	return store.NewErrNotAuthorized("Rename", user, curZid)
}

// DeleteZettel removes the zettel from the store.
func (ps *polStore) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	meta, err := ps.store.GetMeta(ctx, zid)
	if err != nil {
		return err
	}
	user := session.GetUser(ctx)
	if ps.policy.CanDelete(user, meta) {
		return ps.store.DeleteZettel(ctx, zid)
	}
	return store.NewErrNotAuthorized("Delete", user, zid)
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (ps *polStore) Reload(ctx context.Context) error {
	user := session.GetUser(ctx)
	if ps.policy.CanReload(user) {
		return ps.store.Reload(ctx)
	}
	return store.NewErrNotAuthorized("Reload", user, domain.InvalidZettelID)
}
