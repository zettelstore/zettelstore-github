//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package progplace provides zettel that inform the user about the internal Zettelstore state.
package progplace

import (
	"context"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

type (
	zettelGen struct {
		meta    func(domain.ZettelID) *domain.Meta
		content func(*domain.Meta) string
	}

	progPlace struct {
		zettel map[domain.ZettelID]zettelGen
	}
)

var myPlace = progPlace{
	zettel: map[domain.ZettelID]zettelGen{
		domain.ZettelID(1): {genVersionBuildM, genVersionBuildC},
		domain.ZettelID(2): {genVersionHostM, genVersionHostC},
		domain.ZettelID(3): {genVersionOSM, genVersionOSC},
		domain.ZettelID(4): {genVersionGoM, genVersionGoC},
	},
}

// Get returns the one program place.
func Get() place.Place { return &myPlace }

func (pp *progPlace) Next() place.Place { return nil }

// Location returns some information where the place is located.
func (pp *progPlace) Location() string { return "" }

// Start the place. Now all other functions of the place are allowed.
// Starting an already started place is not allowed.
func (pp *progPlace) Start(ctx context.Context) error {
	return nil
}

// Stop the started place. Now only the Start() function is allowed.
func (pp *progPlace) Stop(ctx context.Context) error { return nil }

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
func (pp *progPlace) RegisterChangeObserver(f place.ObserverFunc) {}

func (pp *progPlace) CanCreateZettel(ctx context.Context) bool { return false }

func (pp *progPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	return domain.InvalidZettelID, place.ErrReadOnly
}

// GetZettel retrieves a specific zettel.
func (pp *progPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	if gen, ok := pp.zettel[zid]; ok && gen.meta != nil {
		if meta := gen.meta(zid); meta != nil {
			if genContent := gen.content; genContent != nil {
				return domain.Zettel{Meta: meta, Content: domain.NewContent(genContent(meta))}, nil
			}
			return domain.Zettel{Meta: meta}, nil
		}
	}
	return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
}

// GetMeta retrieves just the meta data of a specific zettel.
func (pp *progPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	if gen, ok := pp.zettel[zid]; ok {
		if genMeta := gen.meta; genMeta != nil {
			if meta := genMeta(zid); meta != nil {
				return meta, nil
			}
		}
	}
	return nil, &place.ErrUnknownID{Zid: zid}
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (pp *progPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) (res []*domain.Meta, err error) {
	hasMatch := place.CreateFilterFunc(f)
	for zid, gen := range pp.zettel {
		if genMeta := gen.meta; genMeta != nil {
			if meta := genMeta(zid); meta != nil && hasMatch(meta) {
				res = append(res, meta)
			}
		}
	}
	return place.ApplySorter(res, s), nil
}

func (pp *progPlace) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool { return false }

func (pp *progPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	return place.ErrReadOnly
}

func (pp *progPlace) CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool { return false }

// DeleteZettel removes the zettel from the place.
func (pp *progPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	return place.ErrReadOnly
}

func (pp *progPlace) CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool {
	_, ok := pp.zettel[zid]
	return !ok
}

// Rename changes the current id to a new id.
func (pp *progPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	return place.ErrReadOnly
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (pp *progPlace) Reload(ctx context.Context) error { return nil }
