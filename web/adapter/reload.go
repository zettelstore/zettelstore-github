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
	"net/http"

	"zettelstore.de/z/encoder"
	"zettelstore.de/z/usecase"
)

// MakeReloadHandler creates a new HTTP handler for the use case "reload".
func MakeReloadHandler(reload usecase.Reload) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := reload.Run(r.Context())
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		format := getFormat(r, encoder.GetDefaultFormat())
		if format == "html" {
			http.Redirect(w, r, urlForList('/'), http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", format2ContentType(format))
		w.WriteHeader(http.StatusNoContent)
	}
}
