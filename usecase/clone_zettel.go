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

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

// CloneZettelPort is the interface used by this use case.
type CloneZettelPort interface {
	// CreateZettel creates a new zettel.
	CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error)
}

// CloneZettel is the data for this use case.
type CloneZettel struct {
	port CloneZettelPort
}

// NewCloneZettel creates a new use case.
func NewCloneZettel(port CloneZettelPort) CloneZettel {
	return CloneZettel{port: port}
}

// Run executes the use case.
func (uc CloneZettel) Run(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	meta := zettel.Meta
	if meta.Zid.IsValid() {
		return meta.Zid, nil // TODO: new error: already exists
	}

	if title, ok := meta.Get(domain.MetaKeyTitle); !ok || title == "" {
		meta.Set(domain.MetaKeyTitle, config.GetDefaultTitle())
	}
	if role, ok := meta.Get(domain.MetaKeyRole); !ok || role == "" {
		meta.Set(domain.MetaKeyRole, config.GetDefaultRole())
	}
	if syntax, ok := meta.Get(domain.MetaKeySyntax); !ok || syntax == "" {
		meta.Set(domain.MetaKeySyntax, config.GetDefaultSyntax())
	}
	meta.YamlSep = config.GetYAMLHeader()

	return uc.port.CreateZettel(ctx, zettel)
}
