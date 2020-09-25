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

// Package chainstore provides a union of connected zettel stores of different
// type.
package chainstore

import (
	"context"
	"errors"
	"strings"
	"sync"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

// chStore implements a chained store.
type chStore struct {
	stores     []store.Store
	observers  []store.ObserverFunc
	mxObserver sync.RWMutex
}

var errEmpty = errors.New("Empty store chain")

// NewStore creates a new chain stores, initialized with given stores.
func NewStore(stores ...store.Store) store.Store {
	cs := new(chStore)
	for _, s := range stores {
		if s != nil {
			cs.stores = append(cs.stores, s)
			s.RegisterChangeObserver(cs.notifyChanged)
		}
	}
	return cs
}

func (cs *chStore) notifyChanged(all bool, zid domain.ZettelID) {
	cs.mxObserver.RLock()
	observers := cs.observers
	cs.mxObserver.RUnlock()
	for _, ob := range observers {
		ob(all, zid)
	}
}

// Location returns some information where the store is located.
// Format is dependent of the store.
func (cs *chStore) Location() string {
	var sb strings.Builder
	for i, s := range cs.stores {
		if i == 0 {
			sb.WriteByte('[')
		} else {
			sb.WriteString(", ")
		}
		sb.WriteString(s.Location())
	}
	sb.WriteByte(']')
	return sb.String()
}

// Start the store. Now all other functions of the store are allowed.
// Starting an already started store is not allowed.
func (cs *chStore) Start(ctx context.Context) error {
	nStores := len(cs.stores)
	if nStores == 0 {
		return errEmpty
	}
	for i, s := range cs.stores {
		if err := s.Start(ctx); err != nil {
			for j := i; j >= 0; j-- {
				cs.stores[j].Stop(ctx)
			}
			return err
		}
	}
	return nil
}

// Stop the started store. Now only the Start() function is allowed.
func (cs *chStore) Stop(ctx context.Context) error {
	nStores := len(cs.stores)
	if nStores == 0 {
		return errEmpty
	}
	var err error
	for i := nStores - 1; i >= 0; i-- {
		if err1 := cs.stores[i].Stop(ctx); err1 != nil && err == nil {
			err = err1
		}
	}
	return err
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (cs *chStore) RegisterChangeObserver(f store.ObserverFunc) {
	cs.mxObserver.Lock()
	cs.observers = append(cs.observers, f)
	cs.mxObserver.Unlock()
}

func (cs *chStore) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	if len(cs.stores) > 0 {
		return cs.stores[0].CreateZettel(ctx, zettel)
	}
	return domain.InvalidZettelID, errEmpty
}

// GetZettel reads the zettel from a file.
func (cs *chStore) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	nStores := len(cs.stores)
	if nStores == 0 {
		return domain.Zettel{}, errEmpty
	}

	for i := 0; i < nStores; i++ {
		zettel, err := cs.stores[i].GetZettel(ctx, zid)
		if err == nil {
			return zettel, nil
		}
		if e, ok := err.(*store.ErrUnknownID); !ok || e.Zid != zid {
			return domain.Zettel{}, err
		}
	}
	return domain.Zettel{}, &store.ErrUnknownID{Zid: zid}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (cs *chStore) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	nStores := len(cs.stores)
	if nStores == 0 {
		return nil, errEmpty
	}

	for i := 0; i < nStores; i++ {
		meta, err := cs.stores[i].GetMeta(ctx, zid)
		if err == nil {
			return meta, nil
		}
		if e, ok := err.(*store.ErrUnknownID); !ok || e.Zid != zid {
			return nil, err
		}
	}
	return nil, &store.ErrUnknownID{Zid: zid}
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (cs *chStore) SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) (res []*domain.Meta, err error) {
	nStores := len(cs.stores)
	if nStores == 0 {
		return nil, errEmpty
	}

	sMetas := make([][]*domain.Meta, 0, nStores)
	hits := 0
	// Could be done in parallel in the future, if needed.
	// Basically, this is a map step
	for i := 0; i < nStores; i++ {
		// No filtering, because of overlay zettel.
		// Sub-stores must order by id, descending. The merge process relies on this.
		metas, err1 := cs.stores[i].SelectMeta(ctx, nil, nil)
		if err1 == nil {
			sMetas = append(sMetas, metas)
			hits += len(metas)
		} else if err == nil {
			err = err1
		}
	}

	// This is the reduce step
	hasMatch := store.CreateFilterFunc(f)
	res = make([]*domain.Meta, 0, hits)
	sPos := make([]int, len(sMetas))
	for {
		maxI := -1
		maxID := int64(-1)
		for i, pos := range sPos {
			if pos < len(sMetas[i]) {
				if zid := int64(sMetas[i][pos].Zid); zid > maxID {
					maxID = zid
					maxI = i
				} else if zid == maxID {
					sPos[i]++
				}
			}
		}
		if maxI < 0 {
			return store.ApplySorter(res, s), nil
		}
		if m := sMetas[maxI][sPos[maxI]]; hasMatch(m) {
			res = append(res, m)
		}
		sPos[maxI]++
	}
}

func (cs *chStore) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	if len(cs.stores) > 0 {
		return cs.stores[0].UpdateZettel(ctx, zettel)
	}
	return errEmpty
}

// Rename changes the current zid to a new zid.
func (cs *chStore) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	if len(cs.stores) == 0 {
		return errEmpty
	}
	for i, s := range cs.stores {
		if err := s.RenameZettel(ctx, curZid, newZid); err != nil {
			if i > 0 {
				return nil
			}
			return err
		}
	}
	return nil
}

// DeleteZettel removes the zettel from the store.
func (cs *chStore) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	if len(cs.stores) > 0 {
		return cs.stores[0].DeleteZettel(ctx, zid)
	}
	return errEmpty
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (cs *chStore) Reload(ctx context.Context) error {
	nStores := len(cs.stores)
	if nStores == 0 {
		return errEmpty
	}
	var err error
	for i := nStores - 1; i >= 0; i-- {
		err1 := cs.stores[i].Reload(ctx)
		if err == nil {
			err = err1
		}
	}
	return err
}
