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

// MakeEditGetZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeEditGetZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
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

		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Edit zettel %q not possible in format %q", id, format), http.StatusNotFound)
			log.Println(err)
			return
		}

		lang := zettel.Meta.GetDefault("lang", config.Config.GetDefaultLang())
		te.renderTemplate(ctx, w, domain.FormTemplateID, formZettelData{
			Meta:    zettel.Meta,
			Lang:    lang,
			Title:   "Edit Zettel",
			Content: zettel.Content.AsString(),
		})
	}
}

// MakeEditSetZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakeEditSetZettelHandler(updateZettel usecase.UpdateZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		zettel, err := parseZettelForm(r, id)
		if err != nil {
			http.Error(w, "Unable to read zettel form", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := updateZettel.Run(r.Context(), zettel); err != nil {
			http.Error(w, fmt.Sprintf("Unable to update zettel %q", id), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(w, r, urlForZettel('h', id), http.StatusFound)
	}
}
