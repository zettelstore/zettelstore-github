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

// MakeGetRenameZettelHandler creates a new HTTP handler to display the HTML rename view of a zettel.
func MakeGetRenameZettelHandler(te *TemplateEngine, getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := domain.ZettelID(r.URL.Path[1:])
		if !id.IsValid() {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		meta, err := getMeta.Run(ctx, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Zettel %q not found", id), http.StatusNotFound)
			log.Println(err)
			return
		}

		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Rename zettel %q not possible in format %q", id, format), http.StatusNotFound)
			log.Println(err)
			return
		}

		te.renderTemplate(ctx, w, domain.RenameTemplateID, struct {
			Title string
			Meta  *domain.Meta
			Lang  string
		}{
			Title: "Rename Zettel " + string(id),
			Meta:  meta,
			Lang:  meta.GetDefault("lang", config.GetDefaultLang()),
		})
	}
}

// MakePostRenameZettelHandler creates a new HTTP handler to rename an existing zettel.
func MakePostRenameZettelHandler(renameZettel usecase.RenameZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		curID := domain.ZettelID(r.URL.Path[1:])
		if !curID.IsValid() {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to read rename zettel form", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if formCurID := domain.ZettelID(r.PostFormValue("curid")); formCurID != curID {
			http.Error(w, "Invalid value for current ID in form", http.StatusBadRequest)
			return
		}
		newID := domain.ZettelID(r.PostFormValue("newid"))
		if !newID.IsValid() {
			http.Error(w, fmt.Sprintf("Invalid new ID %q", newID), http.StatusBadRequest)
			return
		}

		if err := renameZettel.Run(r.Context(), curID, newID); err != nil {
			http.Error(w, fmt.Sprintf("Unable to rename zettel %q", curID), http.StatusBadRequest)
			log.Println(err)
			return
		}
		http.Redirect(w, r, urlFor('h', newID), http.StatusFound)
	}
}
