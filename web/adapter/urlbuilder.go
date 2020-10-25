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
	"net/url"
	"strings"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

type urlQuery struct{ key, val string }
type urlBuilder struct {
	key      byte
	path     []string
	query    []urlQuery
	fragment string
}

func newURLBuilder(key byte) *urlBuilder {
	return &urlBuilder{key: key}
}

func (ub *urlBuilder) Clone() *urlBuilder {
	copy := new(urlBuilder)
	copy.key = ub.key
	if len(ub.path) > 0 {
		copy.path = make([]string, 0, len(ub.path))
	}
	for _, p := range ub.path {
		copy.path = append(copy.path, p)
	}
	if len(ub.query) > 0 {
		copy.query = make([]urlQuery, 0, len(ub.query))
	}
	for _, q := range ub.query {
		copy.query = append(copy.query, q)
	}
	copy.fragment = ub.fragment
	return copy
}

func (ub *urlBuilder) SetZid(zid domain.ZettelID) *urlBuilder {
	if len(ub.path) > 0 {
		panic("Cannot add Zid")
	}
	ub.path = append(ub.path, zid.Format())
	return ub
}

func (ub *urlBuilder) AppendPath(p string) *urlBuilder {
	ub.path = append(ub.path, p)
	return ub
}

func (ub *urlBuilder) AppendQuery(key string, value string) *urlBuilder {
	ub.query = append(ub.query, urlQuery{key, value})
	return ub
}

func (ub *urlBuilder) ClearQuery() *urlBuilder {
	ub.query = nil
	ub.fragment = ""
	return ub
}

func (ub *urlBuilder) SetFragment(s string) *urlBuilder {
	ub.fragment = s
	return ub
}

func (ub *urlBuilder) String() string {
	var sb strings.Builder

	sb.WriteString(config.URLPrefix())
	if ub.key != '/' {
		sb.WriteByte(ub.key)
	}
	for _, p := range ub.path {
		sb.WriteByte('/')
		sb.WriteString(url.PathEscape(p))
	}
	if len(ub.fragment) > 0 {
		sb.WriteByte('#')
		sb.WriteString(ub.fragment)
	}
	for i, q := range ub.query {
		if i == 0 {
			sb.WriteByte('?')
		} else {
			sb.WriteByte('&')
		}
		sb.WriteString(q.key)
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(q.val))
	}
	return sb.String()
}
