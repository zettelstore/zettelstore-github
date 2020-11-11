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
		z := parser.ParseZettel(zettel, syntax)

		langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(z.InhMeta)}
		textTitle, err := formatInlines(z.Title, "text", &langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		metaHeader, err := formatMeta(
			z.InhMeta,
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
		newWindow := true
		htmlContent, err := formatBlocks(
			z.Ast,
			"html",
			&langOption,
			&encoder.StringOption{Key: domain.MetaKeyMarkerExternal, Value: config.GetMarkerExternal()},
			&encoder.BoolOption{Key: "newwindow", Value: newWindow},
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
		canClone := base.CanCreate && !zettel.Content.IsBinary()
		te.renderTemplate(ctx, w, domain.DetailTemplateID, struct {
			baseData
			MetaHeader   template.HTML
			HTMLTitle    template.HTML
			CanWrite     bool
			EditURL      string
			Zid          string
			InfoURL      string
			RoleText     string
			RoleURL      string
			HasTags      bool
			Tags         []simpleLink
			CanClone     bool
			CloneURL     string
			CanNew       bool
			NewURL       string
			HasExtURL    bool
			ExtURL       string
			ExtNewWindow template.HTMLAttr
			Content      template.HTML
		}{
			baseData:     base,
			MetaHeader:   template.HTML(metaHeader),
			HTMLTitle:    template.HTML(htmlTitle),
			CanWrite:     te.canWrite(ctx, user, zettel),
			EditURL:      newURLBuilder('e').SetZid(zid).String(),
			Zid:          zid.Format(),
			InfoURL:      newURLBuilder('i').SetZid(zid).String(),
			RoleText:     roleText,
			RoleURL:      newURLBuilder('h').AppendQuery("role", roleText).String(),
			HasTags:      len(tags) > 0,
			Tags:         tags,
			CanClone:     canClone,
			CloneURL:     newURLBuilder('c').SetZid(zid).String(),
			CanNew:       canClone && roleText == domain.MetaValueRoleNewTemplate,
			NewURL:       newURLBuilder('n').SetZid(zid).String(),
			ExtURL:       extURL,
			HasExtURL:    hasExtURL,
			ExtNewWindow: htmlAttrNewWindow(newWindow && hasExtURL),
			Content:      template.HTML(htmlContent),
		})
	}
}
