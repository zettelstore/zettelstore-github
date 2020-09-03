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
	"zettelstore.de/z/usecase"
)

// MakeGetContentHandler creates a HTTP handler to return the zettel content.
func MakeGetContentHandler(getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		zettel, err := getZettel.Run(r.Context(), zid)
		if err != nil {
			http.Error(w, fmt.Sprintf("Zettel %q not found", zid), http.StatusNotFound)
			log.Println(err)
			return
		}
		syntax := config.GetSyntax(zettel.Meta)
		if contentType, ok := syntaxType[syntax]; ok {
			w.Header().Add("Content-Type", contentType)
		}
		_, err = w.Write(zettel.Content.AsBytes())
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
		}
	}
}

var syntaxType = map[string]string{
	"css":           "text/css; charset=utf-8",
	"gif":           "image/gif",
	"html":          "text/html; charset=utf-8",
	"jpeg":          "image/jpeg",
	"jpg":           "image/jpeg",
	"js":            "text/javascript; charset=utf-8",
	"pdf":           "application/pdf",
	"png":           "image/png",
	"svg":           "image/svg+xml",
	"xml":           "text/xml; charset=utf-8",
	"zmk":           "text/x-zmk; charset=utf-8",
	"plain":         "text/plain; charset=utf-8",
	"text":          "text/plain; charset=utf-8",
	"markdown":      "text/markdown; charset=utf-8",
	"md":            "text/markdown; charset=utf-8",
	"graphviz":      "text/vnd.graphviz; charset=utf-8",
	"blockdiag":     "text/x-blockdiag; charset=utf-8",
	"seqdiag":       "text/x-seqdiag; charset=utf-8",
	"actdiag":       "text/x-actdiag; charset=utf-8",
	"nwdiag":        "text/x-nwdiag; charset=utf-8",
	"plantuml":      "text/x-plantuml; charset=utf-8",
	"template":      "text/x-go-html-template; charset=utf-8",
	"template-html": "text/x-go-html-template; charset=utf-8",
	"template-text": "text/x-go-text-template; charset=utf-8",
}
