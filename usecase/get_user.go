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
	"zettelstore.de/z/store"
)

// Use case: return user identified by meta key ident.
// ---------------------------------------------------

// GetUserPort is the interface used by this use case.
type GetUserPort interface {
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
	SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) ([]*domain.Meta, error)
}

// GetUser is the data for this use case.
type GetUser struct {
	store GetUserPort
}

// NewGetUser creates a new use case.
func NewGetUser(port GetUserPort) GetUser {
	return GetUser{store: port}
}

// Run executes the use case.
func (uc GetUser) Run(ctx context.Context, ident string) (*domain.Meta, error) {
	owner := config.Owner()
	if !owner.IsValid() {
		return nil, nil
	}

	// It is important to try first with the owner. First, because another user
	// could give herself the same ''ident''. Second, in most cases the owner
	// will authenticate.
	identMeta, err := uc.store.GetMeta(ctx, owner)
	if err == nil && identMeta.GetDefault(domain.MetaKeyIdent, "") == ident {
		if role, ok := identMeta.Get(domain.MetaKeyRole); !ok || role != "user" {
			return nil, nil
		}
		return identMeta, nil
	}
	// Owner was not found or has another ident. Try via list search.
	filter := store.Filter{
		Expr: map[string][]string{
			domain.MetaKeyIdent: []string{ident},
			domain.MetaKeyRole:  []string{"user"},
		},
	}
	metaList, err := uc.store.SelectMeta(ctx, &filter, nil)
	if err != nil {
		return nil, err
	}
	if len(metaList) < 1 {
		return nil, nil
	}
	return metaList[len(metaList)-1], nil
}

// Use case: return an user identified by zettel id and assert given ident value.
// ------------------------------------------------------------------------------

// GetUserByZidPort is the interface used by this use case.
type GetUserByZidPort interface {
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
}

// GetUserByZid is the data for this use case.
type GetUserByZid struct {
	store GetUserByZidPort
}

// NewGetUserByZid creates a new use case.
func NewGetUserByZid(port GetUserByZidPort) GetUserByZid {
	return GetUserByZid{store: port}
}

// Run executes the use case.
func (uc GetUserByZid) Run(ctx context.Context, zid domain.ZettelID, ident string) (*domain.Meta, error) {
	userMeta, err := uc.store.GetMeta(ctx, zid)
	if err != nil {
		return nil, err
	}

	if val, ok := userMeta.Get(domain.MetaKeyIdent); !ok || val != ident {
		return nil, nil
	}
	return userMeta, nil
}
