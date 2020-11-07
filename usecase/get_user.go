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
	"zettelstore.de/z/place"
)

// Use case: return user identified by meta key ident.
// ---------------------------------------------------

// GetUserPort is the interface used by this use case.
type GetUserPort interface {
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
	SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error)
}

// GetUser is the data for this use case.
type GetUser struct {
	port GetUserPort
}

// NewGetUser creates a new use case.
func NewGetUser(port GetUserPort) GetUser {
	return GetUser{port: port}
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
	identMeta, err := uc.port.GetMeta(ctx, owner)
	if err == nil && identMeta.GetDefault(domain.MetaKeyUserID, "") == ident {
		if role, ok := identMeta.Get(domain.MetaKeyRole); !ok || role != domain.MetaValueRoleUser {
			return nil, nil
		}
		return identMeta, nil
	}
	// Owner was not found or has another ident. Try via list search.
	filter := place.Filter{
		Expr: map[string][]string{
			domain.MetaKeyRole:   []string{domain.MetaValueRoleUser},
			domain.MetaKeyUserID: []string{ident},
		},
	}
	metaList, err := uc.port.SelectMeta(ctx, &filter, nil)
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
	port GetUserByZidPort
}

// NewGetUserByZid creates a new use case.
func NewGetUserByZid(port GetUserByZidPort) GetUserByZid {
	return GetUserByZid{port: port}
}

// Run executes the use case.
func (uc GetUserByZid) Run(ctx context.Context, zid domain.ZettelID, ident string) (*domain.Meta, error) {
	userMeta, err := uc.port.GetMeta(ctx, zid)
	if err != nil {
		return nil, err
	}

	if val, ok := userMeta.Get(domain.MetaKeyUserID); !ok || val != ident {
		return nil, nil
	}
	return userMeta, nil
}
