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
	"log"
	"net/http"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
)

// MakeListMetaHandler creates a new HTTP handler for the use case "list some zettel".
func MakeListMetaHandler(te *TemplateEngine, listMeta usecase.ListMeta, getMeta usecase.GetMeta, parseZettel usecase.ParseZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filter, sorter := getFilterSorter(q, false)
		metaList, err := listMeta.Run(r.Context(), filter, sorter)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		format := getFormat(r, q, encoder.GetDefaultFormat())
		part := getPart(q, "meta")
		w.Header().Set("Content-Type", format2ContentType(format))
		switch format {
		case "html":
			renderListMetaHTML(w, metaList)
		case "json", "djson":
			renderListMetaJSON(r.Context(), w, metaList, format, part, getMeta, parseZettel)
		case "native", "raw", "text", "zmk":
			http.Error(w, fmt.Sprintf("Zettel list in format %q not yet implemented", format), http.StatusNotImplemented)
			log.Println(format)
		default:
			http.Error(w, fmt.Sprintf("Zettel list not available in format %q", format), http.StatusBadRequest)
		}
	}
}

func renderListMetaHTML(w http.ResponseWriter, metaList []*domain.Meta) {
	buf := encoder.NewBufWriter(w)

	buf.WriteStrings("<html lang=\"", config.GetDefaultLang(), "\">\n<body>\n<ul>\n")
	for _, meta := range metaList {
		title := meta.GetDefault(domain.MetaKeyTitle, "")
		htmlTitle, err := formatInlines(parser.ParseTitle(title), "html")
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		buf.WriteStrings(
			"<li><a href=\"", newURLBuilder('z').SetZid(meta.Zid).AppendQuery("format", "html").String(), "\">",
			htmlTitle, "</a></li>\n")
	}
	buf.WriteString("</ul>\n</body>\n</html>")
	buf.Flush()
}
