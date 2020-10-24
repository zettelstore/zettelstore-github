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

// MakeEditGetZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeEditGetZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		zettel, err := getZettel.Run(ctx, zid)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Edit zettel %q not possible in format %q", zid.Format(), format), http.StatusBadRequest)
			return
		}

		user := session.GetUser(ctx)
		meta := zettel.Meta
		te.renderTemplate(ctx, w, domain.FormTemplateID, formZettelData{
			baseData:      te.makeBaseData(ctx, config.GetLang(meta), "Edit Zettel", user),
			MetaTitle:     meta.GetDefault(domain.MetaKeyTitle, ""),
			MetaTags:      meta.GetDefault(domain.MetaKeyTags, ""),
			MetaRole:      meta.GetDefault(domain.MetaKeyRole, ""),
			MetaSyntax:    meta.GetDefault(domain.MetaKeySyntax, ""),
			MetaPairsRest: meta.PairsRest(),
			IsTextContent: !zettel.Content.IsBinary(),
			Content:       zettel.Content.AsString(),
		})
	}
}

// MakeEditSetZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakeEditSetZettelHandler(updateZettel usecase.UpdateZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		zettel, hasContent, err := parseZettelForm(r, zid)
		if err != nil {
			http.Error(w, "Unable to read zettel form", http.StatusBadRequest)
			return
		}

		if err := updateZettel.Run(r.Context(), zettel, hasContent); err != nil {
			checkUsecaseError(w, err)
			return
		}
		http.Redirect(w, r, newURLBuilder('h').SetZid(zid).String(), http.StatusFound)
	}
}
