//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
		func(u *url.URL, next place.Place) (place.Place, error) {
			return &memPlace{u: u, next: next}, nil
		})
}

type memPlace struct {
	u         *url.URL
	next      place.Place
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

func (mp *memPlace) Next() place.Place { return nil }

func (mp *memPlace) Location() string {
	return mp.u.String()
}

func (mp *memPlace) Start(ctx context.Context) error {
	if mp.next != nil {
		if err := mp.next.Start(ctx); err != nil {
			return err
		}
	}
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
	if mp.next != nil {
		return mp.next.Stop(ctx)
	}
	return nil
}

func (mp *memPlace) RegisterChangeObserver(f place.ObserverFunc) {
	if mp.next != nil {
		mp.next.RegisterChangeObserver(f)
	}
	mp.mx.Lock()
	mp.observers = append(mp.observers, f)
	mp.mx.Unlock()
}

func (mp *memPlace) CanCreateZettel(ctx context.Context) bool {
	return mp.started
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
	if !mp.started {
		mp.mx.RUnlock()
		return domain.Zettel{}, place.ErrStopped
	}
	zettel, ok := mp.zettel[zid]
	mp.mx.RUnlock()
	if !ok {
		if mp.next != nil {
			return mp.next.GetZettel(ctx, zid)
		}
		return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
	}
	return zettel, nil
}

func (mp *memPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	mp.mx.RLock()
	if !mp.started {
		mp.mx.RUnlock()
		return nil, place.ErrStopped
	}
	zettel, ok := mp.zettel[zid]
	mp.mx.RUnlock()
	if !ok {
		if mp.next != nil {
			return mp.next.GetMeta(ctx, zid)
		}
		return nil, &place.ErrUnknownID{Zid: zid}
	}
	return zettel.Meta, nil
}

func (mp *memPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error) {
	mp.mx.RLock()
	if !mp.started {
		mp.mx.RUnlock()
		return nil, place.ErrStopped
	}
	filterFunc := place.CreateFilterFunc(f)
	result := make([]*domain.Meta, 0)
	for _, zettel := range mp.zettel {
		if filterFunc(zettel.Meta) {
			result = append(result, zettel.Meta)
		}
	}
	mp.mx.RUnlock()
	if mp.next != nil {
		other, err := mp.next.SelectMeta(ctx, f, nil)
		if err != nil {
			return nil, err
		}
		return place.MergeSorted(place.ApplySorter(result, nil), other, s), nil
	}
	return place.ApplySorter(result, s), nil
}

func (mp *memPlace) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool {
	return mp.started
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

func (mp *memPlace) CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool {
	mp.mx.RLock()
	defer mp.mx.Unlock()
	if !mp.started {
		return false
	}
	_, ok := mp.zettel[zid]
	return ok || (mp.next != nil && mp.next.CanDeleteZettel(ctx, zid))
}

func (mp *memPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}
	if _, ok := mp.zettel[zid]; !ok {
		if mp.next != nil {
			return mp.next.DeleteZettel(ctx, zid)
		}
		return &place.ErrUnknownID{Zid: zid}
	}
	delete(mp.zettel, zid)
	mp.notifyChanged(false, zid)
	return nil
}

func (mp *memPlace) CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool {
	mp.mx.RLock()
	defer mp.mx.Unlock()
	if !mp.started {
		return false
	}
	_, ok := mp.zettel[zid]
	return ok || (mp.next != nil && mp.next.CanRenameZettel(ctx, zid))
}

func (mp *memPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	mp.mx.Lock()
	defer mp.mx.Unlock()
	if !mp.started {
		return place.ErrStopped
	}
	zettel, ok := mp.zettel[curZid]
	if !ok {
		if mp.next != nil {
			return mp.next.RenameZettel(ctx, curZid, newZid)
		}
		return nil
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
	if mp.next != nil {
		return mp.next.Reload(ctx)
	}
	return nil
}
