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
)

// AuthenticatePort is the interface used by this use case.
type AuthenticatePort interface {
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
func (uc Authenticate) Run(ctx context.Context, ident string, credential string) (bool, error) {
	// Too simple authentication.
	// TODO: more realisitic alogrithm

	if len(ident) == 0 || ident[0] == 'x' {
		return false, nil
	}
	if ident[0] == 'q' && ident != credential {
		return false, nil
	}
	return true, nil
}
