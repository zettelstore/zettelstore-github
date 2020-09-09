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
	"html/template"
	"log"
	"net/http"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/collect"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

type internalReference struct {
	Zid   domain.ZettelID
	Found bool
	Title template.HTML
}

// MakeGetInfoHandler creates a new HTTP handler for the use case "get zettel".
func MakeGetInfoHandler(te *TemplateEngine, getZettel usecase.GetZettel, getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Zettel info not available in format %q", format), http.StatusNotFound)
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
			http.Error(w, fmt.Sprintf("Zettel %q not found", zid.Format()), http.StatusNotFound)
			log.Println(err)
			return
		}
		syntax := r.URL.Query().Get("syntax")
		z, meta := parser.ParseZettel(zettel, syntax)

		langOption := &encoder.StringOption{Key: "lang", Value: config.GetLang(meta)}
		getTitle := func(zid domain.ZettelID) (string, bool) {
			meta, err := getMeta.Run(r.Context(), zid)
			if err != nil {
				return "", false
			}
			astTitle := parser.ParseTitle(meta.GetDefault(domain.MetaKeyTitle, ""))
			title, err := formatInlines(astTitle, "html", langOption)
			if err == nil {
				return title, true
			}
			return "", true
		}
		links, images := collect.References(z)
		intLinks, extLinks := splitIntExtLinks(getTitle, append(links, images...))

		// Render as HTML
		textTitle, err := formatInlines(z.Title, "text", nil, langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		te.renderTemplate(ctx, w, domain.InfoTemplateID, struct {
			Lang     string
			Title    string
			User     userWrapper
			Meta     metaWrapper
			IntLinks []internalReference
			ExtLinks []string
			Formats  []string
		}{
			Lang:     langOption.Value,
			Title:    textTitle, // TODO: merge with site-title?
			User:     wrapUser(session.GetUser(ctx)),
			Meta:     wrapMeta(z.Meta),
			IntLinks: intLinks,
			ExtLinks: extLinks,
			Formats:  encoder.GetFormats(),
		})
	}
}

func splitIntExtLinks(getTitle func(domain.ZettelID) (string, bool), links []*ast.Reference) ([]internalReference, []string) {
	if len(links) == 0 {
		return nil, nil
	}
	intLinks := make([]internalReference, 0, len(links))
	extLinks := make([]string, 0, len(links))
	for _, ref := range links {
		if ref.IsZettel() {
			zid, err := domain.ParseZettelID(ref.Value)
			if err != nil {
				panic(err)
			}
			title, ok := getTitle(zid)
			if len(title) == 0 {
				title = ref.Value
			}
			intLinks = append(intLinks, internalReference{zid, ok, template.HTML(title)})
		} else {
			extLinks = append(extLinks, ref.String())
		}
	}
	return intLinks, extLinks
}
