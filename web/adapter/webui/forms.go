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
	"net/http"
	"strings"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
)

type formZettelData struct {
	baseData
	Heading       template.HTML
	MetaTitle     string
	MetaTags      string
	MetaRole      string
	MetaSyntax    string
	MetaPairsRest []domain.MetaPair
	IsTextContent bool
	Content       string
}

func parseZettelForm(r *http.Request, zid domain.ZettelID) (domain.Zettel, bool, error) {
	err := r.ParseForm()
	if err != nil {
		return domain.Zettel{}, false, err
	}

	var meta *domain.Meta
	if postMeta, ok := trimmedFormValue(r, "meta"); ok {
		meta = domain.NewMetaFromInput(zid, input.NewInput(postMeta))
	} else {
		meta = domain.NewMeta(zid)
	}
	if postTitle, ok := trimmedFormValue(r, "title"); ok {
		meta.Set(domain.MetaKeyTitle, postTitle)
	}
	if postTags, ok := trimmedFormValue(r, "tags"); ok {
		if tags := strings.Fields(postTags); len(tags) > 0 {
			meta.SetList(domain.MetaKeyTags, tags)
		}
	}
	if postRole, ok := trimmedFormValue(r, "role"); ok {
		meta.Set(domain.MetaKeyRole, postRole)
	}
	if postSyntax, ok := trimmedFormValue(r, "syntax"); ok {
		meta.Set(domain.MetaKeySyntax, postSyntax)
	}
	if values, ok := r.PostForm["content"]; ok && len(values) > 0 {
		return domain.Zettel{
			Meta:    meta,
			Content: domain.NewContent(strings.ReplaceAll(strings.TrimSpace(values[0]), "\r\n", "\n")),
		}, true, nil
	}
	return domain.Zettel{
		Meta:    meta,
		Content: domain.NewContent(""),
	}, false, nil
}

func trimmedFormValue(r *http.Request, key string) (string, bool) {
	if values, ok := r.PostForm[key]; ok && len(values) > 0 {
		value := strings.TrimSpace(values[0])
		if len(value) > 0 {
			return value, true
		}
	}
	return "", false
}
