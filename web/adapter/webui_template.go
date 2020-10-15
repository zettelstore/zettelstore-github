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

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/auth/token"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/web/session"
)

type templatePlace interface {
	// CanCreateZettel returns true, if place could possibly create a new zettel.
	CanCreateZettel(ctx context.Context) bool

	// GetZettel retrieves a specific zettel.
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)

	// GetMeta retrieves just the meta data of a specific zettel.
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)

	// CanUpdateZettel returns true, if place could possibly update the given zettel.
	CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool

	// CanDeleteZettel returns true, if place could possibly delete the given zettel.
	CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool

	// CanRenameZettel returns true, if place could possibly rename the given zettel.
	CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool
}

// TemplateEngine is the way to render HTML templates.
type TemplateEngine struct {
	place         templatePlace
	templateCache map[domain.ZettelID]*template.Template
	mxCache       sync.RWMutex
	policy        policy.Policy
}

// NewTemplateEngine creates a new TemplateEngine.
func NewTemplateEngine(p place.Place, pol policy.Policy) *TemplateEngine {
	te := &TemplateEngine{
		place:  p,
		policy: pol,
	}
	te.observe(true, domain.InvalidZettelID)
	p.RegisterChangeObserver(te.observe)
	return te
}

func (te *TemplateEngine) observe(all bool, zid domain.ZettelID) {
	te.mxCache.Lock()
	if all || zid == domain.BaseTemplateID {
		te.templateCache = make(map[domain.ZettelID]*template.Template, len(te.templateCache))
	} else {
		delete(te.templateCache, zid)
	}
	te.mxCache.Unlock()
}

func (te *TemplateEngine) cacheSetTemplate(zid domain.ZettelID, t *template.Template) {
	te.mxCache.Lock()
	te.templateCache[zid] = t
	te.mxCache.Unlock()
}

func (te *TemplateEngine) cacheGetTemplate(zid domain.ZettelID) (*template.Template, bool) {
	te.mxCache.RLock()
	t, ok := te.templateCache[zid]
	te.mxCache.RUnlock()
	return t, ok
}

func urlForList(key byte) string {
	prefix := config.URLPrefix()
	if key == '/' {
		return prefix
	}
	return prefix + string(rune(key))
}

func urlForZettel(key byte, zid domain.ZettelID) string {
	var sb strings.Builder

	sb.WriteString(config.URLPrefix())
	sb.WriteByte(key)
	sb.WriteByte('/')
	sb.WriteString(zid.Format())
	return sb.String()
}

func htmlMetaValue(metaW metaWrapper, key string) template.HTML {
	meta := metaW.original
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
		return template.HTML("<a href=\"" + urlForZettel('h', zid) + "\">" + value + "</a>")

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
	b.WriteString(urlForList('h'))
	b.WriteByte('?')
	b.WriteString(template.URLQueryEscaper(key))
	b.WriteByte('=')
	b.WriteString(template.URLQueryEscaper(value))
	b.WriteString("\">")
	b.WriteString(html.EscapeString(value))
	b.WriteString("</a>")
}

func htmlify(s string) template.HTML {
	return template.HTML(s)
}

func join(sl []string) string {
	return strings.Join(sl, " ")
}

var funcMap = template.FuncMap{
	"urlList":       urlForList,
	"urlZettel":     urlForZettel,
	"htmlMetaValue": htmlMetaValue,
	"HTML":          htmlify,
	"join":          join,
}

func (te *TemplateEngine) canReload(ctx context.Context, user *domain.Meta) bool {
	return te.policy.CanReload(user)
}

func (te *TemplateEngine) canCreate(ctx context.Context, user *domain.Meta) bool {
	meta := domain.NewMeta(domain.InvalidZettelID)
	return te.policy.CanCreate(user, meta) && te.place.CanCreateZettel(ctx)
}

func (te *TemplateEngine) canWrite(ctx context.Context, user *domain.Meta, zettel domain.Zettel) bool {
	return te.policy.CanWrite(user, zettel.Meta, zettel.Meta) && te.place.CanUpdateZettel(ctx, zettel)
}

func (te *TemplateEngine) canRename(ctx context.Context, user *domain.Meta, meta *domain.Meta) bool {
	return te.policy.CanRename(user, meta) && te.place.CanRenameZettel(ctx, meta.Zid)
}

func (te *TemplateEngine) canDelete(ctx context.Context, user *domain.Meta, meta *domain.Meta) bool {
	return te.policy.CanDelete(user, meta) && te.place.CanDeleteZettel(ctx, meta.Zid)
}

func (te *TemplateEngine) getTemplate(ctx context.Context, templateID domain.ZettelID) (*template.Template, error) {
	if t, ok := te.cacheGetTemplate(templateID); ok {
		return t, nil
	}
	baseTemplate, ok := te.cacheGetTemplate(domain.BaseTemplateID)
	if !ok {
		baseTemplateZettel, err := te.place.GetZettel(ctx, domain.BaseTemplateID)
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
	realTemplateZettel, err := te.place.GetZettel(ctx, templateID)
	if err != nil {
		return nil, err
	}
	t, err := baseTemplate.Parse(realTemplateZettel.Content.AsString())
	if err == nil {
		te.cacheSetTemplate(templateID, t)
	}
	return t, err
}

type baseData struct {
	Lang          string
	Version       string
	StylesheetURL template.URL
	Title         string
	HomeURL       template.URL
	ListZettelURL template.URL
	ListRolesURL  template.URL
	ListTagsURL   template.URL
	CanCreate     bool
	NewZettelURL  template.URL
	WithAuth      bool
	UserIsValid   bool
	UserZettelURL template.URL
	UserIdent     string
	UserLogoutURL template.URL
	LoginURL      template.URL
	CanReload     bool
	ReloadURL     template.URL
	SearchURL     template.URL
	FooterHTML    template.HTML
}

func (te *TemplateEngine) makeBaseData(
	ctx context.Context, lang string, title string, user *domain.Meta) baseData {
	var (
		userZettelURL template.URL
		userIdent     string
		userLogoutURL template.URL
	)
	if user != nil {
		userZettelURL = template.URL(urlForZettel('h', user.Zid))
		userIdent = user.GetDefault(domain.MetaKeyIdent, "")
		userLogoutURL = template.URL(urlForZettel('a', user.Zid))
	}
	return baseData{
		Lang:          lang,
		Version:       config.GetVersion().Build,
		StylesheetURL: template.URL(urlForZettel('z', domain.BaseCSSID) + "?_format=raw&_part=content"),
		Title:         title,
		HomeURL:       template.URL(urlForList('/')),
		ListZettelURL: template.URL(urlForList('h')),
		ListRolesURL:  template.URL(urlForZettel('k', 2)),
		ListTagsURL:   template.URL(urlForZettel('k', 3)),
		CanCreate:     te.canCreate(ctx, user),
		NewZettelURL:  template.URL(urlForZettel('n', domain.TemplateZettelID)),
		WithAuth:      config.WithAuth(),
		UserIsValid:   user != nil,
		UserZettelURL: userZettelURL,
		UserIdent:     userIdent,
		UserLogoutURL: userLogoutURL,
		LoginURL:      template.URL(urlForList('a')),
		CanReload:     te.canReload(ctx, user),
		ReloadURL:     template.URL(urlForList('c') + "?_format=html"),
		SearchURL:     template.URL(urlForList('s')),
		FooterHTML:    template.HTML(config.GetFooterHTML()),
	}
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
	if user := session.GetUser(ctx); user != nil {
		htmlLifetime, _ := config.TokenLifetime()
		t, err := token.GetToken(user, htmlLifetime, token.KindHTML)
		if err == nil {
			session.SetToken(w, t, htmlLifetime)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to execute template: %v", err), http.StatusInternalServerError)
		log.Println(err)
	}
}
