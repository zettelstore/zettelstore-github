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

// Package usecase provides (business) use cases for the zettelstore.
package usecase

import (
	"context"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/parser"
)

// ParseZettel is the data for this use case.
type ParseZettel struct {
	getZettel GetZettel
}

// NewParseZettel creates a new use case.
func NewParseZettel(getZettel GetZettel) ParseZettel {
	return ParseZettel{getZettel: getZettel}
}

// Run executes the use case.
func (uc ParseZettel) Run(ctx context.Context, zid domain.ZettelID, syntax string) (*ast.ZettelNode, error) {
	zettel, err := uc.getZettel.Run(ctx, zid)
	if err != nil {
		return nil, err
	}

	return parser.ParseZettel(zettel, syntax), nil
}
