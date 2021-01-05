//-----------------------------------------------------------------------------
// Copyright (c) 2021 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package manager coordinates the various places of a Zettelstore.
package manager

import (
	"context"
	"net/url"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
	"zettelstore.de/z/place"
	"zettelstore.de/z/place/progplace"
)

// Manager is a coordinating place.
type Manager struct {
	placeURIs []url.URL
	place     place.Place
}

// New creates a new managing place.
func New(placeURIs []string) (*Manager, error) {
	place, err := connectPlaces(placeURIs, progplace.Get())
	if err != nil {
		return nil, err
	}
	result := &Manager{
		place: place,
	}
	return result, nil
}

// Helper function to connect to all given places
func connectPlaces(placeURIs []string, lastPlace place.Place) (place.Place, error) {
	if len(placeURIs) == 0 {
		return lastPlace, nil
	}
	next, err := connectPlaces(placeURIs[1:], lastPlace)
	if err != nil {
		return nil, err
	}
	p, err := place.Connect(placeURIs[0], next)
	return p, err
}

// Next returns the next place or nil if there is no next place.
func (mgr *Manager) Next() place.Place { return mgr.place.Next() }

// Location returns some information where the place is located.
func (mgr *Manager) Location() string {
	return mgr.place.Location()
}

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (mgr *Manager) Start(ctx context.Context) error {
	return mgr.place.Start(ctx)
}

// Stop the started place. Now only the Start() function is allowed.
func (mgr *Manager) Stop(ctx context.Context) error {
	return mgr.place.Stop(ctx)
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (mgr *Manager) RegisterChangeObserver(f place.ObserverFunc) {
	mgr.place.RegisterChangeObserver(f)
}

// CanCreateZettel returns true, if place could possibly create a new zettel.
func (mgr *Manager) CanCreateZettel(ctx context.Context) bool {
	return mgr.place.CanCreateZettel(ctx)
}

// CreateZettel creates a new zettel.
func (mgr *Manager) CreateZettel(ctx context.Context, zettel domain.Zettel) (id.Zid, error) {
	return mgr.place.CreateZettel(ctx, zettel)
}

// GetZettel retrieves a specific zettel.
func (mgr *Manager) GetZettel(ctx context.Context, zid id.Zid) (domain.Zettel, error) {
	return mgr.place.GetZettel(ctx, zid)
}

// GetMeta retrieves just the meta data of a specific zettel.
func (mgr *Manager) GetMeta(ctx context.Context, zid id.Zid) (*meta.Meta, error) {
	return mgr.place.GetMeta(ctx, zid)
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (mgr *Manager) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*meta.Meta, error) {
	return mgr.place.SelectMeta(ctx, f, s)
}

// CanUpdateZettel returns true, if place could possibly update the given zettel.
func (mgr *Manager) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool {
	return mgr.place.CanUpdateZettel(ctx, zettel)
}

// UpdateZettel updates an existing zettel.
func (mgr *Manager) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {

	return mgr.place.UpdateZettel(ctx, zettel)
}

// CanRenameZettel returns true, if place could possibly rename the given zettel.
func (mgr *Manager) CanRenameZettel(ctx context.Context, zid id.Zid) bool {
	return mgr.place.CanRenameZettel(ctx, zid)
}

// RenameZettel changes the current zid to a new zid.
func (mgr *Manager) RenameZettel(ctx context.Context, curZid, newZid id.Zid) error {
	return mgr.place.RenameZettel(ctx, curZid, newZid)
}

// CanDeleteZettel returns true, if place could possibly delete the given zettel.
func (mgr *Manager) CanDeleteZettel(ctx context.Context, zid id.Zid) bool {
	return mgr.place.CanDeleteZettel(ctx, zid)
}

// DeleteZettel removes the zettel from the place.
func (mgr *Manager) DeleteZettel(ctx context.Context, zid id.Zid) error {
	return mgr.place.DeleteZettel(ctx, zid)
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (mgr *Manager) Reload(ctx context.Context) error {
	return mgr.place.Reload(ctx)
}
