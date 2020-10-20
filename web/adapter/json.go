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
	"context"
	"fmt"
	"log"
	"net/http"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
)

func writeJSONZettel(ctx context.Context, w http.ResponseWriter, z *ast.Zettel, meta *domain.Meta, format string, part string, getMeta usecase.GetMeta) error {
	var err error
	switch part {
	case "zettel":
		err = writeJSONHeader(w, z.Meta, format)
		if err == nil {
			err = writeJSONMeta(w, z, meta, format)
		}
		if err == nil {
			err = writeJSONContent(ctx, w, z, format, part, getMeta)
		}
	case "meta":
		err = writeJSONHeader(w, z.Meta, format)
		if err == nil {
			err = writeJSONMeta(w, z, z.Meta, format)
		}
	case "content":
		err = writeJSONHeader(w, z.Meta, format)
		if err == nil {
			err = writeJSONContent(ctx, w, z, format, part, getMeta)
		}
	case "id":
		writeJSONHeader(w, z.Meta, format)
	default:
		panic(part)
	}
	if err == nil {
		err = writeJSONFooter(w)
	}
	return err
}

var (
	jsonMetaHeader    = []byte(",\"meta\":")
	jsonContentHeader = []byte(",\"content\":")
	jsonHeader1       = []byte("{\"id\":\"")
	jsonHeader2       = []byte("\",\"url\":\"")
	jsonHeader3       = []byte("?_format=")
	jsonHeader4       = []byte("\"")
	jsonFooter        = []byte("}")
)

func writeJSONHeader(w http.ResponseWriter, meta *domain.Meta, format string) error {
	w.Header().Set("Content-Type", format2ContentType(format))
	_, err := w.Write(jsonHeader1)
	if err == nil {
		_, err = w.Write(meta.Zid.FormatBytes())
	}
	if err == nil {
		_, err = w.Write(jsonHeader2)
	}
	if err == nil {
		_, err = w.Write([]byte(urlForZettel('z', meta.Zid)))
	}
	if err == nil && format != encoder.GetDefaultFormat() {
		_, err = w.Write(jsonHeader3)
		if err == nil {
			_, err = w.Write([]byte(format))
		}
	}
	if err == nil {
		_, err = w.Write(jsonHeader4)
	}
	return err
}

func writeJSONMeta(w http.ResponseWriter, z *ast.Zettel, meta *domain.Meta, format string) error {
	_, err := w.Write(jsonMetaHeader)
	if err == nil {
		err = writeMeta(w, meta, format, &encoder.TitleOption{Inline: z.Title})
	}
	return err
}

func writeJSONContent(ctx context.Context, w http.ResponseWriter, z *ast.Zettel, format string, part string, getMeta usecase.GetMeta) error {
	_, err := w.Write(jsonContentHeader)
	if err == nil {
		err = writeContent(w, z, format,
			&encoder.AdaptLinkOption{Adapter: makeLinkAdapter(ctx, 'z', getMeta, part, format)},
			&encoder.AdaptImageOption{Adapter: makeImageAdapter()},
		)
	}
	return err
}

func writeJSONFooter(w http.ResponseWriter) error {
	_, err := w.Write(jsonFooter)
	return err
}

var (
	jsonListHeader = []byte("{\"list\":[")
	jsonListSep    = []byte{','}
	jsonListFooter = []byte("]}")
)

func renderListMetaJSON(ctx context.Context, w http.ResponseWriter, metaList []*domain.Meta, format string, part string, getMeta usecase.GetMeta, getZettel usecase.GetZettel) {
	var readZettel bool
	switch part {
	case "zettel", "content":
		readZettel = true
	case "meta", "id":
		readZettel = false
	default:
		http.Error(w, fmt.Sprintf("Unknown _part=%v parameter", part), http.StatusBadRequest)
		return
	}
	_, err := w.Write(jsonListHeader)
	for i, meta := range metaList {
		if err != nil {
			break
		}
		if i > 0 {
			_, err = w.Write(jsonListSep)
		}
		if err != nil {
			break
		}
		var z *ast.Zettel
		if readZettel {
			zettel, err1 := getZettel.Run(ctx, meta.Zid)
			if err1 == nil {
				z, meta = parser.ParseZettel(zettel, "")
			} else {
				err = err1
			}
			if err != nil {
				break
			}
		} else {
			z = &ast.Zettel{
				Zid:     meta.Zid,
				Meta:    meta,
				Content: "",
				Title:   parser.ParseTitle(meta.GetDefault(domain.MetaKeyTitle, config.GetDefaultTitle())),
				Ast:     nil,
			}
		}
		err = writeJSONZettel(ctx, w, z, z.Meta, format, part, getMeta)
	}
	if err == nil {
		_, err = w.Write(jsonListFooter)
	}
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Println(err)
	}
}
