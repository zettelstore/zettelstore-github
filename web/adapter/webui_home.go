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

// Package adapter provides handlers for web requests.
package adapter

import (
	"context"
	"net/http"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

type getRootStore interface {
	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
}

// MakeGetRootHandler creates a new HTTP handler to show the root URL.
func MakeGetRootHandler(s getRootStore, startNotFound, startFound http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		startID := config.GetStart()
		if startID.IsValid() {
			if _, err := s.GetMeta(r.Context(), startID); err == nil {
				r.URL.Path = "/" + startID.Format()
				startFound(w, r)
				return
			}
		}
		startNotFound(w, r)
	}
}
