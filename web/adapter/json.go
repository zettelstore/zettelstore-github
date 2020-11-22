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
	"io"
	"log"
	"net/http"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/usecase"
)

func writeJSONZettel(ctx context.Context, w http.ResponseWriter, z *ast.ZettelNode, format string, part string, getMeta usecase.GetMeta) error {
	var err error
	switch part {
	case "zettel":
		err = writeJSONHeader(w, z.Zid, format)
		if err == nil {
			err = writeJSONMeta(w, z, format)
		}
		if err == nil {
			err = writeJSONContent(ctx, w, z, format, part, getMeta)
		}
	case "meta":
		err = writeJSONHeader(w, z.Zid, format)
		if err == nil {
			err = writeJSONMeta(w, z, format)
		}
	case "content":
		err = writeJSONHeader(w, z.Zid, format)
		if err == nil {
			err = writeJSONContent(ctx, w, z, format, part, getMeta)
		}
	case "id":
		writeJSONHeader(w, z.Zid, format)
	default:
		panic(part)
	}
	if err == nil {
		err = writeJSONFooter(w)
	}
	return err
}

var (
	jsonMetaHeader         = []byte(",\"meta\":")
	jsonContentHeaderNJSON = []byte(",\"content\":")
	jsonHeader1            = []byte("{\"id\":\"")
	jsonHeader2            = []byte("\",\"url\":\"")
	jsonHeader3            = []byte("?_format=")
	jsonHeader4            = []byte("\"")
	jsonFooter             = []byte("}")
)

func writeJSONHeader(w http.ResponseWriter, zid domain.ZettelID, format string) error {
	w.Header().Set("Content-Type", format2ContentType(format))
	_, err := w.Write(jsonHeader1)
	if err == nil {
		_, err = w.Write(zid.FormatBytes())
	}
	if err == nil {
		_, err = w.Write(jsonHeader2)
	}
	if err == nil {
		_, err = w.Write([]byte(newURLBuilder('z').SetZid(zid).String()))
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

func writeJSONMeta(w io.Writer, z *ast.ZettelNode, format string) error {
	_, err := w.Write(jsonMetaHeader)
	if err == nil {
		err = writeMeta(w, z.InhMeta, format, &encoder.TitleOption{Inline: z.Title})
	}
	return err
}

func writeJSONContent(ctx context.Context, w io.Writer, z *ast.ZettelNode, format string, part string, getMeta usecase.GetMeta) (err error) {
	if format != "json" {
		_, err = w.Write(jsonContentHeaderNJSON)
	} else {
		_, err = w.Write([]byte{','})
	}
	if err == nil {
		err = writeContent(w, z, format,
			&encoder.AdaptLinkOption{Adapter: makeLinkAdapter(ctx, 'z', getMeta, part, format)},
			&encoder.AdaptImageOption{Adapter: makeImageAdapter()},
		)
	}
	return err
}

func writeJSONFooter(w io.Writer) error {
	_, err := w.Write(jsonFooter)
	return err
}

var (
	jsonListHeader = []byte("{\"list\":[")
	jsonListSep    = []byte{','}
	jsonListFooter = []byte("]}")
)

func renderListMetaJSON(ctx context.Context, w http.ResponseWriter, metaList []*domain.Meta, format string, part string, getMeta usecase.GetMeta, parseZettel usecase.ParseZettel) {
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
		var zn *ast.ZettelNode
		if readZettel {
			z, err1 := parseZettel.Run(ctx, meta.Zid, "")
			if err1 != nil {
				err = err1
				break
			}
			zn = z
		} else {
			zn = &ast.ZettelNode{
				Zettel:  domain.Zettel{Meta: meta, Content: ""},
				Zid:     meta.Zid,
				InhMeta: config.AddDefaultValues(meta),
				Title:   parser.ParseTitle(meta.GetDefault(domain.MetaKeyTitle, config.GetDefaultTitle())),
				Ast:     nil,
			}
		}
		err = writeJSONZettel(ctx, w, zn, format, part, getMeta)
	}
	if err == nil {
		_, err = w.Write(jsonListFooter)
	}
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Println(err)
	}
}
