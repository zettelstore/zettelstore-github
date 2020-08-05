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

// MakeGetBodyHandler creates a new HTTP handler to render the zettel body.
func MakeGetBodyHandler(
	key byte,
	te *TemplateEngine,
	p *parser.Parser,
	getZettel usecase.GetZettel,
	getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := domain.ZettelID(r.URL.Path[1:])
		if !id.IsValid() {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		zettel, err := getZettel.Run(ctx, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Zettel %q not found", id), http.StatusNotFound)
			log.Println(err)
			return
		}
		syntax := r.URL.Query().Get("syntax")
		z := p.ParseZettel(zettel, syntax)

		langOption := &encoder.StringOption{Key: "lang", Value: config.Config.GetDefaultLang()}
		format := getFormat(r, "html")
		w.Header().Set("Content-Type", formatContentType(format))
		err = writeBlocks(w,
			z.Ast,
			format,
			langOption,
			&encoder.StringOption{Key: "material", Value: config.Config.GetIconMaterial()},
			&encoder.AdaptLinkOption{Adapter: makeLinkAdapter(ctx, key, getMeta)},
			&encoder.AdaptImageOption{Adapter: makeImageAdapter(key)},
		)
		if err != nil {
			if err == errNoSuchFormat {
				http.Error(w, fmt.Sprintf("Zettel %q not available in format %q", id, format), http.StatusNotFound)
				log.Println(err, format)
				return
			}
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
		}
	}
}
