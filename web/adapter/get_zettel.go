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

// MakeGetZettelHandler creates a new HTTP handler to return a rendered zettel.
func MakeGetZettelHandler(getZettel usecase.GetZettel, getMeta usecase.GetMeta) http.HandlerFunc {
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
		q := r.URL.Query()
		syntax := q.Get("syntax")
		z := parser.ParseZettel(zettel, syntax)

		format := getFormat(r, q, encoder.GetDefaultFormat())
		part := getPart(q, "zettel")
		switch format {
		case "json", "djson":
			switch part {
			case "zettel", "meta", "content", "id":
			default:
				http.Error(w, fmt.Sprintf("Unknown _part=%v parameter", part), http.StatusBadRequest)
				return
			}
			if err := writeJSONZettel(ctx, w, z, format, part, getMeta); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Println(err)
			}
			return
		}

		langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(z.InhMeta)}
		linkAdapter := encoder.AdaptLinkOption{Adapter: makeLinkAdapter(ctx, 'z', getMeta, part, format)}
		imageAdapter := encoder.AdaptImageOption{Adapter: makeImageAdapter()}

		switch part {
		case "zettel":
			inhMeta := false
			if format != "raw" {
				w.Header().Set("Content-Type", format2ContentType(format))
				inhMeta = true
			}
			enc := encoder.Create(format, &langOption,
				&linkAdapter,
				&imageAdapter,
				&encoder.StringsOption{
					Key: "no-meta",
					Value: []string{
						domain.MetaKeyLang,
					},
				},
			)
			if enc == nil {
				err = errNoSuchFormat
			} else {
				_, err = enc.WriteZettel(w, z, inhMeta)
			}
		case "meta":
			w.Header().Set("Content-Type", format2ContentType(format))
			if format == "raw" {
				err = writeMeta(w, z.Zettel.Meta, format) // Don't write inherited meta data, just the raw
			} else {
				err = writeMeta(w, z.InhMeta, format)
			}
		case "content":
			if format == "raw" {
				if ct, ok := syntax2contentType(config.GetSyntax(z.Zettel.Meta)); ok {
					w.Header().Add("Content-Type", ct)
				}
			} else {
				w.Header().Set("Content-Type", format2ContentType(format))
			}
			err = writeContent(w, z, format,
				&langOption,
				&encoder.StringOption{Key: domain.MetaKeyMarkerExternal, Value: config.GetMarkerExternal()},
				&linkAdapter,
				&imageAdapter,
			)
		default:
			http.Error(w, fmt.Sprintf("Unknown _part=%v parameter", part), http.StatusBadRequest)
			return
		}
		if err != nil {
			if err == errNoSuchFormat {
				http.Error(w, fmt.Sprintf("Zettel %q not available in format %q", zid.Format(), format), http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
		}
	}
}
