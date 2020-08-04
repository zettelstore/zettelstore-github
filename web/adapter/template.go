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
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

type templateStore interface {
	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, id domain.ZettelID) (domain.Zettel, error)

	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, id domain.ZettelID) (*domain.Meta, error)
}

// TemplateEngine is the way to render HTML templates.
type TemplateEngine struct {
	store         templateStore
	templateCache map[domain.ZettelID]*template.Template
	mxCache       sync.RWMutex
}

// NewTemplateEngine creates a new TemplateEngine.
func NewTemplateEngine(s store.Store) *TemplateEngine {
	te := &TemplateEngine{
		store: s,
	}
	te.observe("")
	s.RegisterChangeObserver(te.observe)
	return te
}

func (te *TemplateEngine) observe(id domain.ZettelID) {
	te.mxCache.Lock()
	if len(id) == 0 || id == domain.BaseTemplateID {
		te.templateCache = make(map[domain.ZettelID]*template.Template, len(te.templateCache))
	} else {
		delete(te.templateCache, id)
	}
	te.mxCache.Unlock()
}

func (te *TemplateEngine) cacheSetTemplate(id domain.ZettelID, t *template.Template) {
	te.mxCache.Lock()
	te.templateCache[id] = t
	te.mxCache.Unlock()
}

func (te *TemplateEngine) cacheGetTemplate(id domain.ZettelID) (*template.Template, bool) {
	te.mxCache.RLock()
	t, ok := te.templateCache[id]
	te.mxCache.RUnlock()
	return t, ok
}

func urlFor(key byte, id domain.ZettelID) string {
	var sb strings.Builder

	// TODO: get präfix
	sb.WriteByte('/')
	if key != '/' {
		sb.WriteByte(key)
		sb.WriteByte('/')
		if len(id) > 0 {
			sb.WriteString(string(id))
		}
	}
	return sb.String()
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
		return template.HTML("<a href=\"" + urlFor('h', domain.ZettelID(value)) + "\">" + value + "</a>")

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
	b.WriteString(urlFor('h', ""))
	b.WriteByte('?')
	b.WriteString(template.URLQueryEscaper(key))
	b.WriteByte('=')
	b.WriteString(template.URLQueryEscaper(value))
	b.WriteString("\">")
	b.WriteString(html.EscapeString(value))
	b.WriteString("</a>")
}

var funcMap = template.FuncMap{
	"url":           urlFor,
	"htmlMetaValue": htmlMetaValue,
}

func (te *TemplateEngine) getTemplate(ctx context.Context, templateID domain.ZettelID) (*template.Template, error) {
	if t, ok := te.cacheGetTemplate(templateID); ok {
		return t, nil
	}
	baseTemplate, ok := te.cacheGetTemplate(domain.BaseTemplateID)
	if !ok {
		baseTemplateZettel, err := te.store.GetZettel(ctx, domain.BaseTemplateID)
		if err != nil {
			return nil, err
		}
		baseTemplate, err = template.New("base").Funcs(funcMap).Parse(baseTemplateZettel.Content.AsString())
		if err != nil {
			return nil, err
		}
		te.cacheSetTemplate(domain.BaseTemplateID, baseTemplate)
	}
	baseTemplate, err := baseTemplate.Clone()
	if err != nil {
		return nil, err
	}
	realTemplateZettel, err := te.store.GetZettel(ctx, templateID)
	if err != nil {
		return nil, err
	}
	t, err := baseTemplate.Parse(realTemplateZettel.Content.AsString())
	if err == nil {
		te.cacheSetTemplate(templateID, t)
	}
	return t, err
}

func (te *TemplateEngine) renderTemplate(
	ctx context.Context,
	w http.ResponseWriter,
	templateID domain.ZettelID,
	data interface{}) {

	t, err := te.getTemplate(ctx, templateID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to get template: %v", err), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to execute template: %v", err), http.StatusInternalServerError)
		log.Println(err)
	}
}
