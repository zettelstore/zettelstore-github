//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package webui provides wet-UI handlers for web requests.
package webui

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/session"
)

// MakeGetHTMLZettelHandler creates a new HTTP handler for the use case "get zettel".
func MakeGetHTMLZettelHandler(
	te *TemplateEngine,
	parseZettel usecase.ParseZettel,
	getMeta usecase.GetMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		syntax := r.URL.Query().Get("syntax")
		zn, err := parseZettel.Run(ctx, zid, syntax)
		if err != nil {
			adapter.ReportUsecaseError(w, err)
			return
		}

		metaHeader, err := formatMeta(
			zn.InhMeta,
			"html",
			&encoder.StringsOption{
				Key:   "no-meta",
				Value: []string{domain.MetaKeyTitle, domain.MetaKeyLang},
			},
		)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		langOption := encoder.StringOption{Key: "lang", Value: config.GetLang(zn.InhMeta)}
		htmlTitle, err := adapter.FormatInlines(zn.Title, "html", &langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		textTitle, err := adapter.FormatInlines(zn.Title, "text", &langOption)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		newWindow := true
		htmlContent, err := formatBlocks(
			zn.Ast,
			"html",
			&langOption,
			&encoder.StringOption{Key: domain.MetaKeyMarkerExternal, Value: config.GetMarkerExternal()},
			&encoder.BoolOption{Key: "newwindow", Value: newWindow},
			&encoder.AdaptLinkOption{Adapter: adapter.MakeLinkAdapter(ctx, 'h', getMeta, "", "")},
			&encoder.AdaptImageOption{Adapter: adapter.MakeImageAdapter()},
		)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		user := session.GetUser(ctx)
		roleText := zn.Zettel.Meta.GetDefault(domain.MetaKeyRole, "*")
		tags := buildTagInfos(zn.Zettel.Meta)
		extURL, hasExtURL := zn.Zettel.Meta.Get(domain.MetaKeyURL)
		base := te.makeBaseData(ctx, langOption.Value, textTitle, user)
		canClone := base.CanCreate && !zn.Zettel.Content.IsBinary()
		te.renderTemplate(ctx, w, domain.DetailTemplateID, struct {
			baseData
			MetaHeader   template.HTML
			HTMLTitle    template.HTML
			CanWrite     bool
			EditURL      string
			Zid          string
			InfoURL      string
			RoleText     string
			RoleURL      string
			HasTags      bool
			Tags         []simpleLink
			CanClone     bool
			CloneURL     string
			CanNew       bool
			NewURL       string
			HasExtURL    bool
			ExtURL       string
			ExtNewWindow template.HTMLAttr
			Content      template.HTML
		}{
			baseData:     base,
			MetaHeader:   template.HTML(metaHeader),
			HTMLTitle:    template.HTML(htmlTitle),
			CanWrite:     te.canWrite(ctx, user, zn.Zettel),
			EditURL:      adapter.NewURLBuilder('e').SetZid(zid).String(),
			Zid:          zid.Format(),
			InfoURL:      adapter.NewURLBuilder('i').SetZid(zid).String(),
			RoleText:     roleText,
			RoleURL:      adapter.NewURLBuilder('h').AppendQuery("role", roleText).String(),
			HasTags:      len(tags) > 0,
			Tags:         tags,
			CanClone:     canClone,
			CloneURL:     adapter.NewURLBuilder('c').SetZid(zid).String(),
			CanNew:       canClone && roleText == domain.MetaValueRoleNewTemplate,
			NewURL:       adapter.NewURLBuilder('n').SetZid(zid).String(),
			ExtURL:       extURL,
			HasExtURL:    hasExtURL,
			ExtNewWindow: htmlAttrNewWindow(newWindow && hasExtURL),
			Content:      template.HTML(htmlContent),
		})
	}
}

func formatBlocks(bs ast.BlockSlice, format string, options ...encoder.Option) (string, error) {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return "", adapter.ErrNoSuchFormat
	}

	var content strings.Builder
	_, err := enc.WriteBlocks(&content, bs)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func formatMeta(meta *domain.Meta, format string, options ...encoder.Option) (string, error) {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return "", adapter.ErrNoSuchFormat
	}

	var content strings.Builder
	_, err := enc.WriteMeta(&content, meta)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func buildTagInfos(meta *domain.Meta) []simpleLink {
	var tagInfos []simpleLink
	if tags, ok := meta.GetList(domain.MetaKeyTags); ok {
		tagInfos = make([]simpleLink, 0, len(tags))
		ub := adapter.NewURLBuilder('h')
		for _, t := range tags {
			// Cast to template.HTML is ok, because "t" is a tag name
			// and contains only legal characters by construction.
			tagInfos = append(tagInfos, simpleLink{Text: template.HTML(t), URL: ub.AppendQuery("tags", t).String()})
			ub.ClearQuery()
		}
	}
	return tagInfos
}
