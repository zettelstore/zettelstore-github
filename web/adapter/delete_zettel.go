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

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

type deleteZettelData struct {
	Lang  string
	Title string
	Meta  *domain.Meta
}

// MakeGetDeleteZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetDeleteZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Delete zettel not possible in format %q", format), http.StatusNotFound)
			return
		}

		id := domain.ZettelID(r.URL.Path[1:])
		if !id.IsValid() {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		zettel, err := getZettel.Run(ctx, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Zettel %q not found", id), http.StatusNotFound)
			log.Println(err)
			return
		}

		lang := zettel.Meta.GetDefault(domain.MetaKeyLang, config.GetDefaultLang())
		te.renderTemplate(ctx, w, domain.DeleteTemplateID, deleteZettelData{
			Meta:  zettel.Meta,
			Title: "Delete Zettel " + string(zettel.Meta.ID),
			Lang:  lang,
		})
	}
}

// MakePostDeleteZettelHandler creates a new HTTP handler to delete a zettel.
func MakePostDeleteZettelHandler(deleteZettel usecase.DeleteZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := domain.ZettelID(r.URL.Path[1:])
		if !id.IsValid() {
			http.NotFound(w, r)
			return
		}

		if err := deleteZettel.Run(r.Context(), id); err != nil {
			http.Error(w, fmt.Sprintf("Unable to delete zettel %q", id), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(w, r, urlFor('/', ""), http.StatusFound)
	}
}
