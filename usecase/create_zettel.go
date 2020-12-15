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

	"zettelstore.de/z/config/runtime"
	"zettelstore.de/z/domain"
)

// CreateZettelPort is the interface used by this use case.
type CreateZettelPort interface {
	// CreateZettel creates a new zettel.
	CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error)
}

// CreateZettel is the data for this use case.
type CreateZettel struct {
	port CreateZettelPort
}

// NewCreateZettel creates a new use case.
func NewCreateZettel(port CreateZettelPort) CreateZettel {
	return CreateZettel{port: port}
}

// Run executes the use case.
func (uc CreateZettel) Run(
	ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	meta := zettel.Meta
	if meta.Zid.IsValid() {
		return meta.Zid, nil // TODO: new error: already exists
	}

	if title, ok := meta.Get(domain.MetaKeyTitle); !ok || title == "" {
		meta.Set(domain.MetaKeyTitle, runtime.GetDefaultTitle())
	}
	if role, ok := meta.Get(domain.MetaKeyRole); !ok || role == "" {
		meta.Set(domain.MetaKeyRole, runtime.GetDefaultRole())
	}
	if syntax, ok := meta.Get(domain.MetaKeySyntax); !ok || syntax == "" {
		meta.Set(domain.MetaKeySyntax, runtime.GetDefaultSyntax())
	}
	meta.YamlSep = runtime.GetYAMLHeader()

	return uc.port.CreateZettel(ctx, zettel)
}
