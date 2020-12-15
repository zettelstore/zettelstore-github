//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package webui provides wet-UI handlers for web requests.
package webui

import (
	"fmt"
	"net/http"

	"zettelstore.de/z/config/runtime"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/session"
)

// MakeEditGetZettelHandler creates a new HTTP handler to display the
// HTML edit view of a zettel.
func MakeEditGetZettelHandler(
	te *TemplateEngine, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		zettel, err := getZettel.Run(ctx, zid)
		if err != nil {
			adapter.ReportUsecaseError(w, err)
			return
		}

		if format := adapter.GetFormat(r, r.URL.Query(), "html"); format != "html" {
			http.Error(
				w,
				fmt.Sprintf("Edit zettel %q not possible in format %q", zid.Format(), format),
				http.StatusBadRequest)
			return
		}

		user := session.GetUser(ctx)
		meta := zettel.Meta
		te.renderTemplate(ctx, w, domain.FormTemplateID, formZettelData{
			baseData:      te.makeBaseData(ctx, runtime.GetLang(meta), "Edit Zettel", user),
			MetaTitle:     meta.GetDefault(domain.MetaKeyTitle, ""),
			MetaRole:      meta.GetDefault(domain.MetaKeyRole, ""),
			MetaTags:      meta.GetDefault(domain.MetaKeyTags, ""),
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
			adapter.ReportUsecaseError(w, err)
			return
		}
		http.Redirect(w, r, adapter.NewURLBuilder('h').SetZid(zid).String(), http.StatusFound)
	}
}
