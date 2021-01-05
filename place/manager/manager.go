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
	"log"
	"net/url"
	"sort"
	"strings"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
	"zettelstore.de/z/place"
	"zettelstore.de/z/place/progplace"
)

// Connect returns a handle to the specified place
func Connect(rawURL string, next place.Place) (place.Place, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "dir"
	}
	if create, ok := registry[u.Scheme]; ok {
		return create(u, next)
	}
	return nil, &ErrInvalidScheme{u.Scheme}
}

// ErrInvalidScheme is returned if there is no place with the given scheme
type ErrInvalidScheme struct{ Scheme string }

func (err *ErrInvalidScheme) Error() string { return "Invalid scheme: " + err.Scheme }

type createFunc func(*url.URL, place.Place) (place.Place, error)

var registry = map[string]createFunc{}

// Register the encoder for later retrieval.
func Register(scheme string, create createFunc) {
	if _, ok := registry[scheme]; ok {
		log.Fatalf("Place with scheme %q already registered", scheme)
	}
	registry[scheme] = create
}

// GetSchemes returns all registered scheme, ordered by scheme string.
func GetSchemes() []string {
	result := make([]string, 0, len(registry))
	for scheme := range registry {
		result = append(result, scheme)
	}
	sort.Strings(result)
	return result
}

// Manager is a coordinating place.
type Manager struct {
	placeURIs []url.URL
	place     place.Place
	subplaces []place.Place
}

// New creates a new managing place.
func New(placeURIs []string) (*Manager, error) {
	subplaces := make([]place.Place, 0, 8)
	place, err := connectPlaces(placeURIs, progplace.Get())
	if err != nil {
		return nil, err
	}
	for p := place; p != nil; p = p.Next() {
		subplaces = append(subplaces, p)
	}
	result := &Manager{
		place:     place,
		subplaces: subplaces,
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
	p, err := Connect(placeURIs[0], next)
	return p, err
}

// Next returns the next place or nil if there is no next place.
func (mgr *Manager) Next() place.Place { return mgr.place.Next() }

// Location returns some information where the place is located.
func (mgr *Manager) Location() string {
	if len(mgr.subplaces) < 2 {
		return mgr.subplaces[0].Location()
	}
	var sb strings.Builder
	for i := 0; i < len(mgr.subplaces)-1; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(mgr.subplaces[i].Location())
	}
	return sb.String()
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
