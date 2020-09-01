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
	Meta    *domain.Meta
	Lang    string
	Title   string
	Content string
}

func parseZettelForm(r *http.Request, zid domain.ZettelID) (domain.Zettel, error) {
	err := r.ParseForm()
	if err != nil {
		return domain.Zettel{}, err
	}

	var meta *domain.Meta
	if postMeta := strings.TrimSpace(r.PostFormValue("meta")); postMeta == "" {
		meta = domain.NewMeta(zid)
	} else {
		meta = domain.NewMetaFromInput(zid, input.NewInput(postMeta))
	}
	if postTitle := strings.TrimSpace(r.PostFormValue("title")); postTitle != "" {
		meta.Set(domain.MetaKeyTitle, postTitle)
	}
	if postTags := strings.Fields(r.PostFormValue("tags")); len(postTags) > 0 {
		meta.SetList(domain.MetaKeyTags, postTags)
	}
	if postRole := strings.TrimSpace(r.PostFormValue("role")); postRole != "" {
		meta.Set(domain.MetaKeyRole, postRole)
	}
	if postSyntax := strings.TrimSpace(r.PostFormValue("syntax")); postSyntax != "" {
		meta.Set(domain.MetaKeySyntax, postSyntax)
	}
	return domain.Zettel{
		Meta:    meta,
		Content: domain.NewContent(strings.ReplaceAll(r.PostFormValue("content"), "\r\n", "\n")),
	}, nil
}
