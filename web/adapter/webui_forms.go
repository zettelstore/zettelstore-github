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
	"net/http"
	"strings"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
)

type formZettelData struct {
	baseData
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
	postContent, hasContent := trimmedFormValue(r, "content")
	return domain.Zettel{
		Meta:    meta,
		Content: domain.NewContent(strings.ReplaceAll(postContent, "\r\n", "\n")),
	}, hasContent, nil
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
