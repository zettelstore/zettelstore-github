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
	store     AuthenticatePort
	ucGetUser GetUser
}

// NewAuthenticate creates a new use case.
func NewAuthenticate(port AuthenticatePort) Authenticate {
	return Authenticate{
		store:     port,
		ucGetUser: NewGetUser(port),
	}
}

// Run executes the use case.
func (uc Authenticate) Run(ctx context.Context, ident string, credential string, d time.Duration) ([]byte, error) {
	identMeta, err := uc.ucGetUser.Run(ctx, ident)
	if identMeta == nil || err != nil {
		wait()
		return nil, err
	}

	if cred, ok := identMeta.Get(domain.MetaKeyCred); ok {
		ok, err := auth.CompareHashAndCredential(cred, identMeta.Zid, ident, credential)
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
	wait()
	return nil, nil
}

// wait for same time as if password was checked, to avoid timing hints.
func wait() {
	auth.CompareHashAndCredential(
		"$2a$10$WHcSO3G9afJ3zlOYQR1suuf83bCXED2jmzjti/MH4YH4l2mivDuze", domain.InvalidZettelID, "", "")
}
