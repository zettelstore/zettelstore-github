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
	"sort"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

// ListRolePort is the interface used by this use case.
type ListRolePort interface {
	// SelectMeta returns all zettel meta data that match the selection
	// criteria. The result is ordered by descending zettel id.
	SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error)
}

// ListRole is the data for this use case.
type ListRole struct {
	port ListRolePort
}

// NewListRole creates a new use case.
func NewListRole(port ListRolePort) ListRole {
	return ListRole{port: port}
}

// Run executes the use case.
func (uc ListRole) Run(ctx context.Context) ([]string, error) {
	metas, err := uc.port.SelectMeta(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	roles := make(map[string]bool, 8)
	for _, meta := range metas {
		if role, ok := meta.Get(domain.MetaKeyRole); ok && role != "" {
			roles[role] = true
		}
	}
	result := make([]string, 0, len(roles))
	for role := range roles {
		result = append(result, role)
	}
	sort.Strings(result)
	return result, nil
}
