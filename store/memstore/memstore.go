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

// Package memstore stores zettel volatile in main memory.
package memstore

import (
	"context"
	"net/url"
	"sync"
	"time"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

type memStore struct {
	u         *url.URL
	zettel    map[domain.ZettelID]domain.Zettel
	started   bool
	mx        sync.RWMutex
	observers []store.ObserverFunc
}

// NewStore returns a reference to the one global gostore.
func NewStore(u *url.URL) (store.Store, error) {
	return &memStore{u: u}, nil
}

func (ms *memStore) notifyChanged(all bool, zid domain.ZettelID) {
	for _, ob := range ms.observers {
		ob(all, zid)
	}
}

func (ms *memStore) Location() string {
	return ms.u.String()
}

func (ms *memStore) Start(ctx context.Context) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if ms.started {
		panic("MemStore started twice")
	}
	ms.zettel = make(map[domain.ZettelID]domain.Zettel)
	ms.started = true
	return nil
}

func (ms *memStore) Stop(ctx context.Context) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if !ms.started {
		return store.ErrStopped
	}
	ms.zettel = nil
	ms.started = false
	return nil
}

func (ms *memStore) RegisterChangeObserver(ob store.ObserverFunc) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.observers = append(ms.observers, ob)
}

func (ms *memStore) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if !ms.started {
		return domain.InvalidZettelID, store.ErrStopped
	}

	meta := zettel.Meta.Clone()
	meta.Zid = ms.calcNewZid()
	meta.Freeze()
	zettel.Meta = meta
	ms.zettel[meta.Zid] = zettel
	ms.notifyChanged(false, meta.Zid)
	return meta.Zid, nil
}

func (ms *memStore) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	if !ms.started {
		return domain.Zettel{}, store.ErrStopped
	}
	zettel, ok := ms.zettel[zid]
	if !ok {
		return domain.Zettel{}, &store.ErrUnknownID{Zid: zid}
	}
	return zettel, nil
}

func (ms *memStore) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	if !ms.started {
		return nil, store.ErrStopped
	}
	zettel, ok := ms.zettel[zid]
	if !ok {
		return nil, &store.ErrUnknownID{Zid: zid}
	}
	return zettel.Meta, nil
}

func (ms *memStore) SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	if !ms.started {
		return nil, store.ErrStopped
	}
	filterFunc := store.CreateFilterFunc(f)
	result := make([]*domain.Meta, 0)
	for _, zettel := range ms.zettel {
		if filterFunc(zettel.Meta) {
			result = append(result, zettel.Meta)
		}
	}
	return store.ApplySorter(result, s), nil
}

func (ms *memStore) SetZettel(ctx context.Context, zettel domain.Zettel) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if !ms.started {
		return store.ErrStopped
	}

	meta := zettel.Meta.Clone()
	if !meta.Zid.IsValid() {
		meta.Zid = ms.calcNewZid()
	}
	meta.Freeze()
	zettel.Meta = meta
	ms.zettel[meta.Zid] = zettel
	ms.notifyChanged(false, meta.Zid)
	return nil
}

func (ms *memStore) calcNewZid() domain.ZettelID {
	zid := domain.NewZettelID(false)
	if _, ok := ms.zettel[zid]; !ok {
		return zid
	}
	for {
		zid = domain.NewZettelID(true)
		if _, ok := ms.zettel[zid]; !ok {
			return zid
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (ms *memStore) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if !ms.started {
		return store.ErrStopped
	}
	delete(ms.zettel, zid)
	ms.notifyChanged(false, zid)
	return nil
}

func (ms *memStore) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if !ms.started {
		return store.ErrStopped
	}
	zettel, ok := ms.zettel[curZid]
	if !ok {
		return &store.ErrUnknownID{Zid: curZid}
	}
	_, ok = ms.zettel[newZid]
	if ok {
		return &store.ErrInvalidID{Zid: newZid}
	}
	meta := zettel.Meta.Clone()
	meta.Zid = newZid
	meta.Freeze()
	zettel.Meta = meta
	ms.zettel[newZid] = zettel
	delete(ms.zettel, curZid)
	ms.notifyChanged(false, curZid)
	return nil
}

func (ms *memStore) Reload(ctx context.Context) error {
	if !ms.started {
		return store.ErrStopped
	}
	return nil
}
