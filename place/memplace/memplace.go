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

// Package memplace stores zettel volatile in main memory.
package memplace

import (
	"context"
	"net/url"
	"sync"
	"time"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

func init() {
	place.Register(
		"mem",
		func(u *url.URL) (place.Place, error) {
			return &memPlace{u: u}, nil
		})
}

type memPlace struct {
	u         *url.URL
	zettel    map[domain.ZettelID]domain.Zettel
	started   bool
	mx        sync.RWMutex
	observers []place.ObserverFunc
}

func (mp *memPlace) notifyChanged(all bool, zid domain.ZettelID) {
	for _, ob := range mp.observers {
		ob(all, zid)
	}
}

func (mp *memPlace) Location() string {
	return mp.u.String()
}

func (mp *memPlace) Start(ctx context.Context) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if mp.started {
		panic("memPlace started twice")
	}
	mp.zettel = make(map[domain.ZettelID]domain.Zettel)
	mp.started = true
	return nil
}

func (mp *memPlace) Stop(ctx context.Context) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}
	mp.zettel = nil
	mp.started = false
	return nil
}

func (mp *memPlace) RegisterChangeObserver(ob place.ObserverFunc) {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	mp.observers = append(mp.observers, ob)
}

func (mp *memPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return domain.InvalidZettelID, place.ErrStopped
	}

	meta := zettel.Meta.Clone()
	meta.Zid = mp.calcNewZid()
	meta.Freeze()
	zettel.Meta = meta
	mp.zettel[meta.Zid] = zettel
	mp.notifyChanged(false, meta.Zid)
	return meta.Zid, nil
}

func (mp *memPlace) calcNewZid() domain.ZettelID {
	zid := domain.NewZettelID(false)
	if _, ok := mp.zettel[zid]; !ok {
		return zid
	}
	for {
		zid = domain.NewZettelID(true)
		if _, ok := mp.zettel[zid]; !ok {
			return zid
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (mp *memPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	mp.mx.RLock()
	defer mp.mx.RUnlock()
	if !mp.started {
		return domain.Zettel{}, place.ErrStopped
	}
	zettel, ok := mp.zettel[zid]
	if !ok {
		return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
	}
	return zettel, nil
}

func (mp *memPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	mp.mx.RLock()
	defer mp.mx.RUnlock()
	if !mp.started {
		return nil, place.ErrStopped
	}
	zettel, ok := mp.zettel[zid]
	if !ok {
		return nil, &place.ErrUnknownID{Zid: zid}
	}
	return zettel.Meta, nil
}

func (mp *memPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error) {
	mp.mx.RLock()
	defer mp.mx.RUnlock()
	if !mp.started {
		return nil, place.ErrStopped
	}
	filterFunc := place.CreateFilterFunc(f)
	result := make([]*domain.Meta, 0)
	for _, zettel := range mp.zettel {
		if filterFunc(zettel.Meta) {
			result = append(result, zettel.Meta)
		}
	}
	return place.ApplySorter(result, s), nil
}

func (mp *memPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}

	meta := zettel.Meta.Clone()
	if !meta.Zid.IsValid() {
		return &place.ErrInvalidID{Zid: meta.Zid}
	}
	meta.Freeze()
	zettel.Meta = meta
	mp.zettel[meta.Zid] = zettel
	mp.notifyChanged(false, meta.Zid)
	return nil
}

func (mp *memPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}
	delete(mp.zettel, zid)
	mp.notifyChanged(false, zid)
	return nil
}

func (mp *memPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}
	zettel, ok := mp.zettel[curZid]
	if !ok {
		return &place.ErrUnknownID{Zid: curZid}
	}
	_, ok = mp.zettel[newZid]
	if ok {
		return &place.ErrInvalidID{Zid: newZid}
	}
	meta := zettel.Meta.Clone()
	meta.Zid = newZid
	meta.Freeze()
	zettel.Meta = meta
	mp.zettel[newZid] = zettel
	delete(mp.zettel, curZid)
	mp.notifyChanged(false, curZid)
	return nil
}

func (mp *memPlace) Reload(ctx context.Context) error {
	if !mp.started {
		return place.ErrStopped
	}
	return nil
}
