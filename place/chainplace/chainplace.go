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

// Package chainplace provides a union of connected zettel places of different
// type.
package chainplace

import (
	"context"
	"errors"
	"strings"
	"sync"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

// chPlace implements a chained place.
type chPlace struct {
	places     []place.Place
	observers  []place.ObserverFunc
	mxObserver sync.RWMutex
}

var errEmpty = errors.New("Empty place chain")

// NewPlace creates a new chain places, initialized with given places.
func NewPlace(places ...place.Place) place.Place {
	cp := new(chPlace)
	for _, p := range places {
		if p != nil {
			cp.places = append(cp.places, p)
			p.RegisterChangeObserver(cp.notifyChanged)
		}
	}
	return cp
}

func (cp *chPlace) notifyChanged(all bool, zid domain.ZettelID) {
	cp.mxObserver.RLock()
	observers := cp.observers
	cp.mxObserver.RUnlock()
	for _, ob := range observers {
		ob(all, zid)
	}
}

// Location returns some information where the place is located.
// Format is dependent of the place.
func (cp *chPlace) Location() string {
	var sb strings.Builder
	for i, p := range cp.places {
		if i == 0 {
			sb.WriteByte('[')
		} else {
			sb.WriteString(", ")
		}
		sb.WriteString(p.Location())
	}
	sb.WriteByte(']')
	return sb.String()
}

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (cp *chPlace) Start(ctx context.Context) error {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return errEmpty
	}
	for i, p := range cp.places {
		if err := p.Start(ctx); err != nil {
			for j := i; j >= 0; j-- {
				cp.places[j].Stop(ctx)
			}
			return err
		}
	}
	return nil
}

// Stop the started place. Now only the Start() function is allowed.
func (cp *chPlace) Stop(ctx context.Context) error {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return errEmpty
	}
	var err error
	for i := nPlaces - 1; i >= 0; i-- {
		if err1 := cp.places[i].Stop(ctx); err1 != nil && err == nil {
			err = err1
		}
	}
	return err
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (cp *chPlace) RegisterChangeObserver(f place.ObserverFunc) {
	cp.mxObserver.Lock()
	cp.observers = append(cp.observers, f)
	cp.mxObserver.Unlock()
}

func (cp *chPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	if len(cp.places) > 0 {
		return cp.places[0].CreateZettel(ctx, zettel)
	}
	return domain.InvalidZettelID, errEmpty
}

// GetZettel reads the zettel from a file.
func (cp *chPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return domain.Zettel{}, errEmpty
	}

	for i := 0; i < nPlaces; i++ {
		zettel, err := cp.places[i].GetZettel(ctx, zid)
		if err == nil {
			return zettel, nil
		}
		if e, ok := err.(*place.ErrUnknownID); !ok || e.Zid != zid {
			return domain.Zettel{}, err
		}
	}
	return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (cp *chPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return nil, errEmpty
	}

	for i := 0; i < nPlaces; i++ {
		meta, err := cp.places[i].GetMeta(ctx, zid)
		if err == nil {
			return meta, nil
		}
		if e, ok := err.(*place.ErrUnknownID); !ok || e.Zid != zid {
			return nil, err
		}
	}
	return nil, &place.ErrUnknownID{Zid: zid}
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (cp *chPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) (res []*domain.Meta, err error) {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return nil, errEmpty
	}

	pMetas := make([][]*domain.Meta, 0, nPlaces)
	hits := 0
	// Could be done in parallel in the future, if needed.
	// Basically, this is a map step
	for i := 0; i < nPlaces; i++ {
		// No filtering, because of overlay zettel.
		// Sub-places must order by id, descending. The merge process relies on this.
		metas, err1 := cp.places[i].SelectMeta(ctx, nil, nil)
		if err1 == nil {
			pMetas = append(pMetas, metas)
			hits += len(metas)
		} else if err == nil {
			err = err1
		}
	}

	// This is the reduce step
	hasMatch := place.CreateFilterFunc(f)
	res = make([]*domain.Meta, 0, hits)
	pPos := make([]int, len(pMetas))
	for {
		maxI := -1
		maxID := int64(-1)
		for i, pos := range pPos {
			if pos < len(pMetas[i]) {
				if zid := int64(pMetas[i][pos].Zid); zid > maxID {
					maxID = zid
					maxI = i
				} else if zid == maxID {
					pPos[i]++
				}
			}
		}
		if maxI < 0 {
			return place.ApplySorter(res, s), nil
		}
		if m := pMetas[maxI][pPos[maxI]]; hasMatch(m) {
			res = append(res, m)
		}
		pPos[maxI]++
	}
}

func (cp *chPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	if len(cp.places) > 0 {
		return cp.places[0].UpdateZettel(ctx, zettel)
	}
	return errEmpty
}

// Rename changes the current zid to a new zid.
func (cp *chPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	if len(cp.places) == 0 {
		return errEmpty
	}
	for i, p := range cp.places {
		if err := p.RenameZettel(ctx, curZid, newZid); err != nil {
			if i > 0 {
				return nil
			}
			return err
		}
	}
	return nil
}

// DeleteZettel removes the zettel from the place.
func (cp *chPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	if len(cp.places) > 0 {
		return cp.places[0].DeleteZettel(ctx, zid)
	}
	return errEmpty
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (cp *chPlace) Reload(ctx context.Context) error {
	nPlaces := len(cp.places)
	if nPlaces == 0 {
		return errEmpty
	}
	var err error
	for i := nPlaces - 1; i >= 0; i-- {
		err1 := cp.places[i].Reload(ctx)
		if err == nil {
			err = err1
		}
	}
	return err
}
