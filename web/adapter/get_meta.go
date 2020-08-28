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
	"fmt"
	"log"
	"net/http"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

// MakeGetMetaHandler creates a new HTTP handler for the use case "get content".
func MakeGetMetaHandler(getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		meta, err := getMeta.Run(r.Context(), id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Meta for zettel %q not found", id), http.StatusNotFound)
			log.Println(err)
			return
		}

		if format := getFormat(r, "raw"); format != "raw" {
			w.Header().Set("Content-Type", formatContentType(format))
			err = writeMeta(w, meta, format)
			if err == errNoSuchFormat {
				http.Error(w, fmt.Sprintf("Meta data for zettel %q not available in format %q", id, format), http.StatusNotFound)
				log.Println(err, format)
				return
			}
			if err != nil {
				log.Println(err)
			}
			return
		}
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		meta.Write(w)
	}
}
