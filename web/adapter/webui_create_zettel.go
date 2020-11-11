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

	"zettelstore.de/z/input"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakeGetCloneZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetCloneZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel, cloneZettel usecase.CloneZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if origZettel, ok := getOrigZettel(w, r, getZettel, "Clone"); ok {
			renderZettelForm(w, r, te, cloneZettel.Run(origZettel), "Clone Zettel", template.HTML("Clone Zettel"))
		}
	}
}

// MakeGetNewZettelHandler creates a new HTTP handler to display the HTML edit view of a zettel.
func MakeGetNewZettelHandler(te *TemplateEngine, getZettel usecase.GetZettel, newZettel usecase.NewZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if origZettel, ok := getOrigZettel(w, r, getZettel, "New"); ok {
			meta := origZettel.Meta
			title := parser.ParseInlines(input.NewInput(config.GetTitle(meta)), "zmk")
			langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(meta)}
			textTitle, err := formatInlines(title, "text", &langOption)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			htmlTitle, err := formatInlines(title, "html", &langOption)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			renderZettelForm(w, r, te, newZettel.Run(origZettel), textTitle, template.HTML(htmlTitle))
		}
	}
}

func getOrigZettel(w http.ResponseWriter, r *http.Request, getZettel usecase.GetZettel, op string) (domain.Zettel, bool) {
	if format := getFormat(r, r.URL.Query(), "html"); format != "html" {
		http.Error(w, fmt.Sprintf("%v zettel not possible in format %q", op, format), http.StatusBadRequest)
		return domain.Zettel{}, false
	}
	zid, err := domain.ParseZettelID(r.URL.Path[1:])
	if err != nil {
		http.NotFound(w, r)
		return domain.Zettel{}, false
	}
	origZettel, err := getZettel.Run(r.Context(), zid)
	if err != nil {
		http.NotFound(w, r)
		return domain.Zettel{}, false
	}
	return origZettel, true
}

func renderZettelForm(w http.ResponseWriter, r *http.Request, te *TemplateEngine, zettel domain.Zettel, title string, heading template.HTML) {
	ctx := r.Context()
	user := session.GetUser(ctx)
	meta := zettel.Meta
	te.renderTemplate(r.Context(), w, domain.FormTemplateID, formZettelData{
		baseData:      te.makeBaseData(ctx, config.GetLang(meta), title, user),
		Heading:       heading,
		MetaTitle:     config.GetTitle(meta),
		MetaTags:      meta.GetDefault(domain.MetaKeyTags, ""),
		MetaRole:      config.GetRole(meta),
		MetaSyntax:    config.GetSyntax(meta),
		MetaPairsRest: meta.PairsRest(),
		IsTextContent: !zettel.Content.IsBinary(),
		Content:       zettel.Content.AsString(),
	})
}

// MakePostCreateZettelHandler creates a new HTTP handler to store content of an existing zettel.
func MakePostCreateZettelHandler(createZettel usecase.CreateZettel) http.HandlerFunc {
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

		if newZid, err := createZettel.Run(r.Context(), zettel); err != nil {
			checkUsecaseError(w, err)
		} else {
			http.Redirect(w, r, newURLBuilder('h').SetZid(newZid).String(), http.StatusFound)
		}
	}
}
