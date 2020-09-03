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
	"log"
	"net/http"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

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

		metas, err := buildHTMLMetaList(metaList)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		te.renderTemplate(r.Context(), w, domain.ListTemplateID, struct {
			Key   byte
			Lang  string
			Title string
			Metas []metaInfo
		}{
			Key:   key,
			Lang:  config.GetDefaultLang(),
			Title: config.GetSiteName(),
			Metas: metas,
		})

	}
}
