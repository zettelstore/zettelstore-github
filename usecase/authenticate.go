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
	"time"

	"zettelstore.de/z/auth"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

// AuthenticatePort is the interface used by this use case.
type AuthenticatePort interface {
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
	SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error)
}

// Authenticate is the data for this use case.
type Authenticate struct {
	store AuthenticatePort
}

// NewAuthenticate creates a new use case.
func NewAuthenticate(port AuthenticatePort) Authenticate {
	return Authenticate{store: port}
}

// Run executes the use case.
func (uc Authenticate) Run(ctx context.Context, ident string, credential string, d time.Duration) ([]byte, error) {
	owner := config.GetOwner()
	if !owner.IsValid() {
		return nil, nil
	}

	// It is important to try first with the owner. First, because another user
	// could give herself the same ''ident''. Second, in most cases the owner
	// will authenticate.
	identMeta, err := uc.store.GetMeta(ctx, owner)

	if err != nil || identMeta.GetDefault(domain.MetaKeyIdent, "") != ident {
		// Owner was not found or has another ident. Try via list search.
		filter := store.Filter{
			Expr: map[string][]string{
				"ident": []string{ident},
			},
		}
		metaList, err := uc.store.SelectMeta(ctx, &filter, nil)
		if err != nil {
			return nil, err
		}
		if len(metaList) < 1 {
			return nil, nil
		}
		identMeta = metaList[len(metaList)-1]
	}

	if role, ok := identMeta.Get(domain.MetaKeyRole); !ok || role != "user" {
		return nil, nil
	}
	if cred, ok := identMeta.Get(domain.MetaKeyCred); ok {
		ok, err := auth.CompareHashAndCredential(cred, credential)
		if err != nil {
			return nil, err
		}
		if ok {
			token, err := auth.GetToken(identMeta, d)
			if err != nil {
				return nil, err
			}
			return token, nil
		}
		return nil, nil
	}
	return nil, nil
}
