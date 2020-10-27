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
	"net/http"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakeGetCloneZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetCloneZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, r.URL.Query(), "html"); format != "html" {
			http.Error(w, fmt.Sprintf("New zettel not possible in format %q", format), http.StatusBadRequest)
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

		user := session.GetUser(ctx)
		meta := zettel.Meta
		te.renderTemplate(r.Context(), w, domain.FormTemplateID, formZettelData{
			baseData:      te.makeBaseData(ctx, config.GetLang(meta), "Clone Zettel", user),
			MetaTitle:     config.GetTitle(meta),
			MetaTags:      meta.GetDefault(domain.MetaKeyTags, ""),
			MetaRole:      config.GetRole(meta),
			MetaSyntax:    config.GetSyntax(meta),
			MetaPairsRest: meta.PairsRest(),
			IsTextContent: !zettel.Content.IsBinary(),
			Content:       zettel.Content.AsString(),
		})
	}
}

// MakePostCloneZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakePostCloneZettelHandler(cloneZettel usecase.CloneZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zettel, hasContent, err := parseZettelForm(r, domain.InvalidZettelID)
		if err != nil {
			http.Error(w, "Unable to read form data", http.StatusBadRequest)
			return
		}
		if !hasContent {
			http.Error(w, "Content is missing", http.StatusBadRequest)
			return
		}

		if newZid, err := cloneZettel.Run(r.Context(), zettel); err != nil {
			checkUsecaseError(w, err)
		} else {
			http.Redirect(w, r, newURLBuilder('h').SetZid(newZid).String(), http.StatusFound)
		}
	}
}