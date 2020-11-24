//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package usecase provides (business) use cases for the zettelstore.
package usecase

import (
	"context"

	"zettelstore.de/z/domain"
)

// RenameZettelPort is the interface used by this use case.
type RenameZettelPort interface {
	// Rename changes the current id to a new id.
	RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error
}

// RenameZettel is the data for this use case.
type RenameZettel struct {
	port RenameZettelPort
}

// NewRenameZettel creates a new use case.
func NewRenameZettel(port RenameZettelPort) RenameZettel {
	return RenameZettel{port: port}
}

// Run executes the use case.
func (uc RenameZettel) Run(ctx context.Context, curID, newID domain.ZettelID) error {
	return uc.port.RenameZettel(ctx, curID, newID)
}
