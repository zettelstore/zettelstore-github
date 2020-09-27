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

// Package stock allows to get zettel without reading it from a place.
package stock

import (
	"context"
	"sync"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

// Place is a place that is used by a stock.
type Place interface {
	// RegisterChangeObserver registers an observer that will be notified
	// if all or one zettel are found to be changed.
	RegisterChangeObserver(ob place.ObserverFunc)

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)
}

// Stock allow to get subscribed zettel without reading it from a place.
type Stock interface {
	Subscribe(zid domain.ZettelID) error
	GetZettel(zid domain.ZettelID) domain.Zettel
	GetMeta(zid domain.ZettelID) *domain.Meta
}

// NewStock creates a new stock that operates on the given place.
func NewStock(place Place) Stock {
	//RegisterChangeObserver(func(domain.ZettelID))
	stock := &defaultStock{
		place: place,
		subs:  make(map[domain.ZettelID]domain.Zettel),
	}
	place.RegisterChangeObserver(stock.observe)
	return stock
}

type defaultStock struct {
	place  Place
	subs   map[domain.ZettelID]domain.Zettel
	mxSubs sync.RWMutex
}

// observe tracks all changes the place signals.
func (s *defaultStock) observe(all bool, zid domain.ZettelID) {
	if !all {
		s.mxSubs.RLock()
		defer s.mxSubs.RUnlock()
		if _, found := s.subs[zid]; found {
			go func() {
				s.mxSubs.Lock()
				defer s.mxSubs.Unlock()
				s.update(zid)
			}()
		}
		return
	}

	go func() {
		s.mxSubs.Lock()
		defer s.mxSubs.Unlock()
		for zid := range s.subs {
			s.update(zid)
		}
	}()
}

func (s *defaultStock) update(zid domain.ZettelID) {
	if zettel, err := s.place.GetZettel(context.Background(), zid); err == nil {
		s.subs[zid] = zettel
		return
	}
}

// Subscribe adds a zettel to the stock.
func (s *defaultStock) Subscribe(zid domain.ZettelID) error {
	s.mxSubs.Lock()
	defer s.mxSubs.Unlock()
	if _, found := s.subs[zid]; found {
		return nil
	}
	zettel, err := s.place.GetZettel(context.Background(), zid)
	if err != nil {
		return err
	}
	s.subs[zid] = zettel
	return nil
}

// GetZettel returns the zettel with the given zid, if in stock, else an empty zettel
func (s *defaultStock) GetZettel(zid domain.ZettelID) domain.Zettel {
	s.mxSubs.RLock()
	defer s.mxSubs.RUnlock()
	return s.subs[zid]
}

// GetZettel returns the zettel Meta with the given zid, if in stock, else nil.
func (s *defaultStock) GetMeta(zid domain.ZettelID) *domain.Meta {
	s.mxSubs.RLock()
	zettel := s.subs[zid]
	s.mxSubs.RUnlock()
	return zettel.Meta
}
