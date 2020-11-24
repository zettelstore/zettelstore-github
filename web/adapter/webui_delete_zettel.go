//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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

// MakeGetDeleteZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetDeleteZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, r.URL.Query(), "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Delete zettel not possible in format %q", format), http.StatusBadRequest)
			return
		}

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

		user := session.GetUser(ctx)
		meta := zettel.Meta
		te.renderTemplate(ctx, w, domain.DeleteTemplateID, struct {
			baseData
			Zid       string
			MetaPairs []domain.MetaPair
		}{
			baseData:  te.makeBaseData(ctx, config.GetLang(meta), "Delete Zettel "+meta.Zid.Format(), user),
			Zid:       zid.Format(),
			MetaPairs: meta.Pairs(),
		})
	}
}

// MakePostDeleteZettelHandler creates a new HTTP handler to delete a zettel.
func MakePostDeleteZettelHandler(deleteZettel usecase.DeleteZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := deleteZettel.Run(r.Context(), zid); err != nil {
			checkUsecaseError(w, err)
			return
		}
		http.Redirect(w, r, newURLBuilder('/').String(), http.StatusFound)
	}
}
