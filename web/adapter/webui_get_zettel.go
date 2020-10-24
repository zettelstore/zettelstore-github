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
	"html/template"
	"log"
	"net/http"
	"net/url"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakeGetHTMLZettelHandler creates a new HTTP handler for the use case "get zettel".
func MakeGetHTMLZettelHandler(
	te *TemplateEngine,
	getZettel usecase.GetZettel,
	getMeta usecase.GetMeta) http.HandlerFunc {
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
		syntax := r.URL.Query().Get("syntax")
		z, meta := parser.ParseZettel(zettel, syntax)

		langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(meta)}
		textTitle, err := formatInlines(z.Title, "text", &langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		metaHeader, err := formatMeta(
			meta,
			"html",
			&encoder.StringsOption{
				Key:   "no-meta",
				Value: []string{domain.MetaKeyTitle, domain.MetaKeyLang},
			},
		)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		htmlTitle, err := formatInlines(z.Title, "html", &langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		htmlContent, err := formatBlocks(
			z.Ast,
			"html",
			&langOption,
			&encoder.StringOption{Key: "material", Value: config.GetIconMaterial()},
			&encoder.BoolOption{Key: "newwindow", Value: true},
			&encoder.AdaptLinkOption{Adapter: makeLinkAdapter(ctx, 'h', getMeta, "", "")},
			&encoder.AdaptImageOption{Adapter: makeImageAdapter()},
		)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		user := session.GetUser(ctx)
		roleText := z.Meta.GetDefault(domain.MetaKeyRole, "*")
		tags := buildTagInfos(z.Meta)
		extURL, hasExtURL := z.Meta.Get(domain.MetaKeyURL)
		base := te.makeBaseData(ctx, langOption.Value, textTitle, user)
		te.renderTemplate(ctx, w, domain.DetailTemplateID, struct {
			baseData
			MetaHeader template.HTML
			HTMLTitle  template.HTML
			CanWrite   bool
			EditURL    string
			Zid        string
			InfoURL    string
			RoleText   string
			RoleURL    string
			HasTags    bool
			Tags       []metaTagInfo
			CanClone   bool
			CloneURL   string
			HasExtURL  bool
			ExtURL     string
			Content    template.HTML
		}{
			baseData:   base,
			MetaHeader: template.HTML(metaHeader),
			HTMLTitle:  template.HTML(htmlTitle),
			CanWrite:   te.canWrite(ctx, user, zettel),
			EditURL:    urlForZettel('e', zid),
			Zid:        zid.Format(),
			InfoURL:    urlForZettel('i', zid),
			RoleText:   roleText,
			RoleURL:    urlForList('h') + "?role=" + url.QueryEscape(roleText),
			HasTags:    len(tags) > 0,
			Tags:       tags,
			CanClone:   base.CanCreate && !zettel.Content.IsBinary(),
			CloneURL:   urlForZettel('n', zid),
			ExtURL:     extURL,
			HasExtURL:  hasExtURL,
			Content:    template.HTML(htmlContent),
		})
	}
}
