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
	"zettelstore.de/z/web/session"
)

// MakeGetNewZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetNewZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("New zettel not possible in format %q", format), http.StatusNotFound)
			return
		}
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()
		oldZettel, err := getZettel.Run(ctx, zid)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		zettel := &domain.Zettel{Meta: oldZettel.Meta.Clone(), Content: oldZettel.Content}

		te.renderTemplate(r.Context(), w, domain.FormTemplateID, formZettelData{
			Lang:    config.GetLang(zettel.Meta),
			Title:   "New Zettel",
			User:    wrapUser(session.GetUser(ctx)),
			Meta:    wrapMeta(zettel.Meta),
			Content: zettel.Content.AsString(),
		})
	}
}

// MakePostNewZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakePostNewZettelHandler(newZettel usecase.NewZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zettel, err := parseZettelForm(r, domain.InvalidZettelID)
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
		http.Redirect(w, r, urlForZettel('h', zettel.Meta.Zid), http.StatusFound)
	}
}
