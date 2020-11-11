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
	"html/template"
	"log"
	"net/http"
	"sync"

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/auth/token"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/place"
	"zettelstore.de/z/web/session"
)

type templatePlace interface {
	CanCreateZettel(ctx context.Context) bool
	GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error)
	GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error)
	SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) ([]*domain.Meta, error)
	CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool
	CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool
	CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool
}

// TemplateEngine is the way to render HTML templates.
type TemplateEngine struct {
	place         templatePlace
	templateCache map[domain.ZettelID]*template.Template
	mxCache       sync.RWMutex
	policy        policy.Policy

	version       string
	stylesheetURL string
	homeURL       string
	listZettelURL string
	listRolesURL  string
	listTagsURL   string
	withAuth      bool
	loginURL      string
	reloadURL     string
	searchURL     string
}

// NewTemplateEngine creates a new TemplateEngine.
func NewTemplateEngine(p place.Place, pol policy.Policy) *TemplateEngine {
	te := &TemplateEngine{
		place:  p,
		policy: pol,

		version:       config.GetVersion().Build,
		stylesheetURL: newURLBuilder('z').SetZid(domain.BaseCSSID).AppendQuery("_format", "raw").AppendQuery("_part", "content").String(),
		homeURL:       newURLBuilder('/').String(),
		listZettelURL: newURLBuilder('h').String(),
		listRolesURL:  newURLBuilder('k').SetZid(2).String(),
		listTagsURL:   newURLBuilder('k').SetZid(3).String(),
		withAuth:      config.WithAuth(),
		loginURL:      newURLBuilder('a').String(),
		reloadURL:     newURLBuilder('c').AppendQuery("_format", "html").String(),
		searchURL:     newURLBuilder('s').String(),
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
		baseTemplate, err = template.New("base").Parse(baseTemplateZettel.Content.AsString())
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

type simpleLink struct {
	Text template.HTML
	URL  string
}

type baseData struct {
	Lang           string
	Version        string
	StylesheetURL  string
	Title          string
	HomeURL        string
	ListZettelURL  string
	ListRolesURL   string
	ListTagsURL    string
	CanCreate      bool
	NewZettelURL   string
	NewZettelLinks []simpleLink
	WithAuth       bool
	UserIsValid    bool
	UserZettelURL  string
	UserIdent      string
	UserLogoutURL  string
	LoginURL       string
	CanReload      bool
	ReloadURL      string
	SearchURL      string
	FooterHTML     template.HTML
}

func (te *TemplateEngine) makeBaseData(
	ctx context.Context, lang string, title string, user *domain.Meta) baseData {
	var (
		newZettelLinks []simpleLink
		userZettelURL  string
		userIdent      string
		userLogoutURL  string
	)
	canCreate := te.canCreate(ctx, user)
	if canCreate {
		newZettelLinks = te.fetchNewTemplates(ctx, user)
	}
	userIsValid := user != nil
	if userIsValid {
		userZettelURL = newURLBuilder('h').SetZid(user.Zid).String()
		userIdent = user.GetDefault(domain.MetaKeyUserID, "")
		userLogoutURL = newURLBuilder('a').SetZid(user.Zid).String()
	}

	return baseData{
		Lang:           lang,
		Version:        te.version,
		StylesheetURL:  te.stylesheetURL,
		Title:          title,
		HomeURL:        te.homeURL,
		ListZettelURL:  te.listZettelURL,
		ListRolesURL:   te.listRolesURL,
		ListTagsURL:    te.listTagsURL,
		CanCreate:      canCreate,
		NewZettelLinks: newZettelLinks,
		WithAuth:       te.withAuth,
		UserIsValid:    userIsValid,
		UserZettelURL:  userZettelURL,
		UserIdent:      userIdent,
		UserLogoutURL:  userLogoutURL,
		LoginURL:       te.loginURL,
		CanReload:      te.policy.CanReload(user),
		ReloadURL:      te.reloadURL,
		SearchURL:      te.searchURL,
		FooterHTML:     template.HTML(config.GetFooterHTML()),
	}
}

// htmlAttrNewWindow eturns HTML attribute string for opening a link in a new window.
// If hasURL is false an empty string is returned.
func htmlAttrNewWindow(hasURL bool) template.HTMLAttr {
	if hasURL {
		return " target=\"_blank\" ref=\"noopener noreferrer\""
	}
	return ""
}

var templatePlaceFilter = &place.Filter{
	Expr: place.FilterExpr{
		domain.MetaKeyRole: []string{domain.MetaValueRoleNewTemplate},
	},
}

var templatePlaceSorter = &place.Sorter{
	Order:      "id",
	Descending: false,
	Offset:     -1,
	Limit:      31, // Just to be one the safe side...
}

func (te *TemplateEngine) fetchNewTemplates(ctx context.Context, user *domain.Meta) []simpleLink {
	templateList, err := te.place.SelectMeta(ctx, templatePlaceFilter, templatePlaceSorter)
	if err != nil {
		return nil
	}
	result := make([]simpleLink, 0, len(templateList))
	for _, meta := range templateList {
		if te.policy.CanRead(user, meta) {
			title := config.GetTitle(meta)
			langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(meta)}
			astTitle := parser.ParseInlines(input.NewInput(config.GetTitle(meta)), "zmk")
			menuTitle, err := formatInlines(astTitle, "html", &langOption)
			if err != nil {
				menuTitle, err = formatInlines(astTitle, "text", &langOption)
				if err != nil {
					menuTitle = title
				}
			}
			result = append(result, simpleLink{
				Text: template.HTML(menuTitle),
				URL:  newURLBuilder('n').SetZid(meta.Zid).String(),
			})
		}
	}
	return result
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
