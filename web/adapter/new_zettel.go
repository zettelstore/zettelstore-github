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

// MakeGetNewZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetNewZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("New zettel not possible in format %q", format), http.StatusNotFound)
			return
		}
		ctx := r.Context()
		var zettel *domain.Zettel
		id := domain.ZettelID(r.URL.Path[1:])
		if !id.IsValid() {
			http.NotFound(w, r)
			return
		}
		oldZettel, err := getZettel.Run(ctx, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		zettel = &domain.Zettel{Meta: oldZettel.Meta.Clone(), Content: oldZettel.Content}

		lang := zettel.Meta.GetDefault(domain.MetaKeyLang, config.Config.GetDefaultLang())
		te.renderTemplate(r.Context(), w, domain.FormTemplateID, formZettelData{
			Meta:    zettel.Meta,
			Lang:    lang,
			Title:   "New Zettel",
			Content: zettel.Content.AsString(),
		})
	}
}

// MakePostNewZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakePostNewZettelHandler(newZettel usecase.NewZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zettel, err := parseZettelForm(r, "")
		if err != nil {
			http.Error(w, "Unable to read form data", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := newZettel.Run(r.Context(), zettel); err != nil {
			http.Error(w, "Unable to create new zettel", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(w, r, urlFor('h', zettel.Meta.ID), http.StatusFound)
	}
}
