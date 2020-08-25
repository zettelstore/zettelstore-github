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
)

type metaInfo struct {
	Meta  *domain.Meta
	Title template.HTML
}

// MakeListHTMLMetaHandler creates a HTTP handler for rendering the list of zettel as HTML.
func MakeListHTMLMetaHandler(key byte, te *TemplateEngine, listMeta usecase.ListMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, sorter := getFilterSorter(r)
		metaList, err := listMeta.Run(r.Context(), filter, sorter)
		if err != nil {
			http.Error(w, "Zettel store not operational", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		langOption := &encoder.StringOption{Key: "lang", Value: config.Config.GetDefaultLang()}
		metas := make([]metaInfo, 0, len(metaList))
		for _, meta := range metaList {
			title, _ := meta.Get(domain.MetaKeyTitle)
			htmlTitle, err := formatInlines(parser.ParseTitle(title), "html", langOption)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			metas = append(metas, metaInfo{meta, template.HTML(htmlTitle)})
		}
		te.renderTemplate(r.Context(), w, domain.ListTemplateID, struct {
			Key   byte
			Lang  string
			Title string
			Metas []metaInfo
		}{
			Key:   key,
			Lang:  langOption.Value,
			Title: config.Config.GetSiteName(),
			Metas: metas,
		})

	}
}
