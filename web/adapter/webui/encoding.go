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
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/web/adapter"
)

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

type metaInfo struct {
	Title template.HTML
	URL   string
	Tags  []simpleLink
}

// buildHTMLMetaList builds a zettel list based on a meta list for HTML rendering.
func buildHTMLMetaList(metaList []*domain.Meta) ([]metaInfo, error) {
	defaultLang := config.GetDefaultLang()
	langOption := encoder.StringOption{Key: "lang", Value: ""}
	metas := make([]metaInfo, 0, len(metaList))
	for _, meta := range metaList {
		if lang, ok := meta.Get(domain.MetaKeyLang); ok {
			langOption.Value = lang
		} else {
			langOption.Value = defaultLang
		}
		title, _ := meta.Get(domain.MetaKeyTitle)
		htmlTitle, err := adapter.FormatInlines(parser.ParseTitle(title), "html", &langOption)
		if err != nil {
			return nil, err
		}
		metas = append(metas, metaInfo{
			Title: template.HTML(htmlTitle),
			URL:   adapter.NewURLBuilder('h').SetZid(meta.Zid).String(),
			Tags:  buildTagInfos(meta),
		})
	}
	return metas, nil
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
