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

// Package gostore stores zettel inside the executable.
package gostore

import (
	"context"
	"errors"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

type goHeader map[string]string

func makeMeta(id domain.ZettelID, h goHeader) *domain.Meta {
	m := domain.NewMeta(id)
	for k, v := range h {
		m.Set(k, v)
	}
	m.Freeze()
	return m
}

type goZettel struct {
	header  goHeader
	content domain.Content
}

type goStore struct {
	name   string
	zettel map[domain.ZettelID]goZettel
}

// NewStore returns a reference to the one global gostore.
func NewStore() store.Store {
	return &goData
}

// SetParentStore is called when the store is part of a bigger store.
func (gs *goStore) SetParentStore(parent store.Store) {}

// Location returns some information where the store is located.
func (gs *goStore) Location() string {
	return gs.name
}

// Start the store. Now all other functions of the store are allowed.
// Starting an already started store is not allowed.
func (gs *goStore) Start(ctx context.Context) error {
	return nil
}

// Stop the started store. Now only the Start() function is allowed.
func (gs *goStore) Stop(ctx context.Context) error {
	return nil
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed. If the id is empty, all zettel are
// possibly changed.
func (gs *goStore) RegisterChangeObserver(f store.ObserverFunc) {
	// This store never changes anything. So ignore the registration.
}

// GetZettel retrieves a specific zettel.
func (gs *goStore) GetZettel(ctx context.Context, id domain.ZettelID) (domain.Zettel, error) {
	if z, ok := gs.zettel[id]; ok {
		return domain.Zettel{Meta: makeMeta(id, z.header), Content: z.content}, nil
	}
	return domain.Zettel{}, &store.ErrUnknownID{ID: id}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (gs *goStore) GetMeta(ctx context.Context, id domain.ZettelID) (*domain.Meta, error) {
	if z, ok := gs.zettel[id]; ok {
		return makeMeta(id, z.header), nil
	}
	return nil, &store.ErrUnknownID{ID: id}
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (gs *goStore) SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) (res []*domain.Meta, err error) {
	hasMatch := store.CreateFilterFunc(f)
	for id, zettel := range gs.zettel {
		meta := makeMeta(id, zettel.header)
		if hasMatch(meta) {
			res = append(res, meta)
		}
	}
	return store.ApplySorter(res, s), nil
}

var errReadOnly = errors.New("Read-only store")

// SetZettel updates a zettel.
func (gs *goStore) SetZettel(ctx context.Context, zettel domain.Zettel) error {
	return errReadOnly
}

// DeleteZettel removes the zettel from the store.
func (gs *goStore) DeleteZettel(ctx context.Context, id domain.ZettelID) error {
	return errReadOnly
}

// Rename changes the current ID to a new ID.
func (gs *goStore) RenameZettel(ctx context.Context, curID, newID domain.ZettelID) error {
	return errReadOnly
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (gs *goStore) Reload(ctx context.Context) error { return nil }
