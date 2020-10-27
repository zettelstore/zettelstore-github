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
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/collect"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/place"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

type metaDataInfo struct {
	Key   string
	Value template.HTML
}

type zettelReference struct {
	Zid    domain.ZettelID
	Title  template.HTML
	HasURL bool
	URL    string
}

type matrixElement struct {
	Text   string
	HasURL bool
	URL    string
}

// MakeGetInfoHandler creates a new HTTP handler for the use case "get zettel".
func MakeGetInfoHandler(te *TemplateEngine, getZettel usecase.GetZettel, getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if format := getFormat(r, q, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Zettel info not available in format %q", format), http.StatusBadRequest)
			return
		}

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
		syntax := q.Get("syntax")
		z, meta := parser.ParseZettel(zettel, syntax)

		langOption := &encoder.StringOption{Key: "lang", Value: config.GetLang(meta)}
		getTitle := func(zid domain.ZettelID) (string, int) {
			meta, err := getMeta.Run(r.Context(), zid)
			if err != nil {
				if place.IsErrNotAllowed(err) {
					return "", -1
				}
				return "", 0
			}
			astTitle := parser.ParseTitle(meta.GetDefault(domain.MetaKeyTitle, ""))
			title, err := formatInlines(astTitle, "html", langOption)
			if err == nil {
				return title, 1
			}
			return "", 1
		}
		links, images := collect.References(z)
		zetLinks, locLinks, extLinks := splitIntExtLinks(getTitle, append(links, images...))

		// Render as HTML
		textTitle, err := formatInlines(z.Title, "text", nil, langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		user := session.GetUser(ctx)
		pairs := z.Meta.Pairs()
		metaData := make([]metaDataInfo, 0, len(pairs))
		for _, p := range pairs {
			metaData = append(metaData, metaDataInfo{p.Key, htmlMetaValue(z.Meta, p.Key)})
		}
		formats := encoder.GetFormats()
		defFormat := encoder.GetDefaultFormat()
		parts := []string{"zettel", "meta", "content"}
		matrix := make([][]matrixElement, 0, len(parts))
		u := newURLBuilder('z').SetZid(zid)
		for _, part := range parts {
			row := make([]matrixElement, 0, len(formats)+1)
			row = append(row, matrixElement{part, false, ""})
			for _, format := range formats {
				u.AppendQuery("_part", part)
				if format != defFormat {
					u.AppendQuery("_format", format)
				}
				row = append(row, matrixElement{format, true, u.String()})
				u.ClearQuery()
			}
			matrix = append(matrix, row)
		}
		base := te.makeBaseData(ctx, langOption.Value, textTitle, user)
		te.renderTemplate(ctx, w, domain.InfoTemplateID, struct {
			baseData
			Zid         string
			WebURL      string
			CanWrite    bool
			EditURL     string
			CanClone    bool
			CloneURL    string
			CanRename   bool
			RenameURL   string
			CanDelete   bool
			DeleteURL   string
			MetaData    []metaDataInfo
			HasLinks    bool
			HasZetLinks bool
			ZetLinks    []zettelReference
			HasLocLinks bool
			LocLinks    []string
			HasExtLinks bool
			ExtLinks    []string
			Matrix      [][]matrixElement
		}{
			baseData:    base,
			Zid:         zid.Format(),
			WebURL:      newURLBuilder('h').SetZid(zid).String(),
			CanWrite:    te.canWrite(ctx, user, zettel),
			EditURL:     newURLBuilder('e').SetZid(zid).String(),
			CanClone:    base.CanCreate && !zettel.Content.IsBinary(),
			CloneURL:    newURLBuilder('c').SetZid(zid).String(),
			CanRename:   te.canRename(ctx, user, zettel.Meta),
			RenameURL:   newURLBuilder('r').SetZid(zid).String(),
			CanDelete:   te.canDelete(ctx, user, zettel.Meta),
			DeleteURL:   newURLBuilder('d').SetZid(zid).String(),
			MetaData:    metaData,
			HasLinks:    len(zetLinks)+len(extLinks)+len(locLinks) > 0,
			HasZetLinks: len(zetLinks) > 0,
			ZetLinks:    zetLinks,
			HasLocLinks: len(locLinks) > 0,
			LocLinks:    locLinks,
			HasExtLinks: len(extLinks) > 0,
			ExtLinks:    extLinks,
			Matrix:      matrix,
		})
	}
}

func htmlMetaValue(meta *domain.Meta, key string) template.HTML {
	switch meta.Type(key) {
	case domain.MetaTypeBool:
		var b strings.Builder
		if meta.GetBool(key) {
			writeLink(&b, key, "True")
		} else {
			writeLink(&b, key, "False")
		}
		return template.HTML(b.String())

	case domain.MetaTypeID:
		value, _ := meta.Get(key)
		zid, err := domain.ParseZettelID(value)
		if err != nil {
			return template.HTML(value)
		}
		return template.HTML("<a href=\"" + newURLBuilder('h').SetZid(zid).String() + "\">" + value + "</a>")

	case domain.MetaTypeTagSet, domain.MetaTypeWordSet:
		values, _ := meta.GetList(key)
		var b strings.Builder
		for i, tag := range values {
			if i > 0 {
				b.WriteByte(' ')
			}
			writeLink(&b, key, tag)
		}
		return template.HTML(b.String())

	case domain.MetaTypeURL:
		value, _ := meta.Get(key)
		url, err := url.Parse(value)
		if err != nil {
			return template.HTML(html.EscapeString(value))
		}
		return template.HTML("<a href=\"" + url.String() + "\">" + html.EscapeString(value) + "</a>")

	case domain.MetaTypeWord:
		value, _ := meta.Get(key)
		var b strings.Builder
		writeLink(&b, key, value)
		return template.HTML(b.String())

	default:
		value, _ := meta.Get(key)
		return template.HTML(html.EscapeString(value))
	}
}

func writeLink(b *strings.Builder, key, value string) {
	b.WriteString("<a href=\"")
	b.WriteString(newURLBuilder('h').String())
	b.WriteByte('?')
	b.WriteString(template.URLQueryEscaper(key))
	b.WriteByte('=')
	b.WriteString(template.URLQueryEscaper(value))
	b.WriteString("\">")
	b.WriteString(html.EscapeString(value))
	b.WriteString("</a>")
}

func splitIntExtLinks(getTitle func(domain.ZettelID) (string, int), links []*ast.Reference) (zetLinks []zettelReference, locLinks []string, extLinks []string) {
	if len(links) == 0 {
		return nil, nil, nil
	}
	for _, ref := range links {
		if ref.IsZettel() {
			zid, err := domain.ParseZettelID(ref.URL.Path)
			if err != nil {
				panic(err)
			}
			title, found := getTitle(zid)
			if found >= 0 {
				if len(title) == 0 {
					title = ref.Value
				}
				var u string
				if found == 1 {
					ub := newURLBuilder('h').SetZid(zid)
					if fragment := ref.URL.EscapedFragment(); len(fragment) > 0 {
						ub.SetFragment(fragment)
					}
					u = ub.String()
				}
				zetLinks = append(zetLinks, zettelReference{zid, template.HTML(title), len(u) > 0, u})
			}
		} else if ref.IsExternal() {
			extLinks = append(extLinks, ref.String())
		} else {
			locLinks = append(locLinks, ref.String())
		}
	}
	return zetLinks, locLinks, extLinks
}
