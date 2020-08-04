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

// Package stock allows to get zettel without reading it from the store.
package stock

import (
	"context"
	"sync"

	"zettelstore.de/z/domain"
)

// Store is a store that is used by a stock.
type Store interface {
	// RegisterChangeObserver registers an observer that will be notified
	// if a zettel was found to be changed. If the id is empty, all zettel are
	// possibly changed.
	RegisterChangeObserver(func(domain.ZettelID))

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, id domain.ZettelID) (domain.Zettel, error)
}

// Stock allow to get subscribed zettel without reading it from a store.
type Stock interface {
	Subscribe(id domain.ZettelID) error
	GetZettel(id domain.ZettelID) domain.Zettel
	GetMeta(id domain.ZettelID) *domain.Meta
}

// NewStock creates a new stock that operates on the given store.
func NewStock(store Store) Stock {
	//RegisterChangeObserver(func(domain.ZettelID))
	stock := &defaultStock{
		store: store,
		subs:  make(map[domain.ZettelID]domain.Zettel),
	}
	store.RegisterChangeObserver(stock.observe)
	return stock
}

type defaultStock struct {
	store  Store
	subs   map[domain.ZettelID]domain.Zettel
	mxSubs sync.RWMutex
}

// observe tracks all changes the store signals.
func (s *defaultStock) observe(id domain.ZettelID) {
	if id != "" {
		s.mxSubs.RLock()
		defer s.mxSubs.RUnlock()
		if _, found := s.subs[id]; found {
			go func() {
				s.mxSubs.Lock()
				defer s.mxSubs.Unlock()
				s.update(id)
			}()
		}
		return
	}

	go func() {
		s.mxSubs.Lock()
		defer s.mxSubs.Unlock()
		for id := range s.subs {
			s.update(id)
		}
	}()
}

func (s *defaultStock) update(id domain.ZettelID) {
	if zettel, err := s.store.GetZettel(context.Background(), id); err == nil {
		s.subs[id] = zettel
		return
	}
}

// Subscribe adds a zettel to the stock.
func (s *defaultStock) Subscribe(id domain.ZettelID) error {
	s.mxSubs.Lock()
	defer s.mxSubs.Unlock()
	if _, found := s.subs[id]; found {
		return nil
	}
	zettel, err := s.store.GetZettel(context.Background(), id)
	if err != nil {
		return err
	}
	s.subs[id] = zettel
	return nil
}

// GetZettel returns the zettel with the given id, if in stock, else an empty zettel
func (s *defaultStock) GetZettel(id domain.ZettelID) domain.Zettel {
	s.mxSubs.RLock()
	defer s.mxSubs.RUnlock()
	return s.subs[id]
}

// GetZettel returns the zettel Meta with the given id, if in stock, else nil.
func (s *defaultStock) GetMeta(id domain.ZettelID) *domain.Meta {
	s.mxSubs.RLock()
	zettel := s.subs[id]
	s.mxSubs.RUnlock()
	return zettel.Meta
}
