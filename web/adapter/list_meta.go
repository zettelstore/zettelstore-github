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
	"zettelstore.de/z/encoder/jsonenc"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
)

// MakeListMetaHandler creates a new HTTP handler for the use case "list some zettel".
func MakeListMetaHandler(key byte, te *TemplateEngine, p *parser.Parser, listMeta usecase.ListMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, sorter := getFilterSorter(r)
		metaList, err := listMeta.Run(r.Context(), filter, sorter)
		if err != nil {
			http.Error(w, "Zettel store not operational", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		format := getFormat(r, "html")
		w.Header().Set("Content-Type", formatContentType(format))
		switch format {
		case "html":
			renderListMetaHTML(w, key, metaList, p)
		case "json":
			renderListMetaJSON(w, metaList, p)
		default:
			http.Error(w, fmt.Sprintf("Zettel list not available in format %q", format), http.StatusNotFound)
			log.Println(err, format)
		}
	}
}

func renderListMetaHTML(w http.ResponseWriter, key byte, metaList []*domain.Meta, p *parser.Parser) {
	buf := encoder.NewBufWriter(w)

	buf.WriteString("<html lang=\"")
	buf.WriteString(config.Config.GetDefaultLang())
	buf.WriteString("\">\n<body>\n<ul>\n")
	for _, meta := range metaList {
		title := meta.GetDefault(domain.MetaKeyTitle, "")
		htmlTitle, err := formatInlines(p.ParseTitle(meta.ID, input.NewInput(title)), "html")
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		buf.WriteString("<li><a href=\"")
		buf.WriteString(urlFor(key, meta.ID))
		buf.WriteString("\">")
		buf.WriteString(htmlTitle)
		buf.WriteString("</a></li>\n")
	}
	buf.WriteString("</ul>\n</body>\n</html>")
	buf.Flush()
}

func renderListMetaJSON(w http.ResponseWriter, metaList []*domain.Meta, p *parser.Parser) {
	buf := encoder.NewBufWriter(w)

	buf.WriteString("{\"list\":[")
	for i, meta := range metaList {
		title := meta.GetDefault(domain.MetaKeyTitle, "")
		jsonTitle, err := formatInlines(p.ParseTitle(meta.ID, input.NewInput(title)), "json")
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString("{\"id\":\"")
		buf.WriteString(string(meta.ID))
		buf.WriteString("\",\"url\":\"")
		buf.WriteString(urlFor('z', meta.ID))
		buf.WriteString("\",\"meta\":{\"title\":")
		buf.WriteString(jsonTitle)
		if syntax, ok := meta.Get(domain.MetaKeySyntax); ok {
			buf.WriteString(",\"syntax\":\"")
			buf.Write(jsonenc.Escape(syntax))
			buf.WriteByte('"')
		}
		if tags, ok := meta.GetList(domain.MetaKeyTags); ok {
			buf.WriteString(",\"tags\":[")
			for j, tag := range tags {
				if j > 0 {
					buf.WriteByte(',')
				}
				buf.WriteByte('"')
				buf.Write(jsonenc.Escape(tag))
				buf.WriteByte('"')
			}
			buf.WriteByte(']')
		}
		if role, ok := meta.Get(domain.MetaKeyRole); ok {
			buf.WriteString(",\"role\":\"")
			buf.Write(jsonenc.Escape(role))
			buf.WriteByte('"')
		}
		if pairs := meta.PairsRest(); len(pairs) > 0 {
			buf.WriteByte(',')
			for j, p := range pairs {
				if j > 0 {
					buf.WriteByte(',')
				}
				buf.WriteByte('"')
				buf.Write(jsonenc.Escape(p.Key))
				buf.WriteString("\":\"")
				buf.Write(jsonenc.Escape(p.Value))
				buf.WriteByte('"')
			}
		}
		buf.WriteString("}}")
	}
	buf.WriteString("]}")
	buf.Flush()
}
