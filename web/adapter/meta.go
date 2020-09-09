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
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

// meta is a wrapper around `domain.Meta` that will be shown on HTML templates.
type metaWrapper struct {
	original *domain.Meta
}

func wrapMeta(original *domain.Meta) metaWrapper {
	return metaWrapper{original}
}

// Zid returns the zettel ID of the wrapped user.
func (m metaWrapper) Zid() domain.ZettelID {
	return m.original.Zid
}

// GetTitle returns the title value or the given default value.
func (m metaWrapper) GetTitle(defaultValue string) string {
	return m.original.GetDefault(domain.MetaKeyTitle, defaultValue)
}

// GetTags returns the list of tags.
func (m metaWrapper) GetTags() []string {
	if tags, ok := m.original.GetList(domain.MetaKeyTags); ok && len(tags) > 0 {
		return tags
	}
	return nil
}

// GetRole returns the role value.
func (m metaWrapper) GetRole(defaultValue string) string {
	return m.original.GetDefault(domain.MetaKeyRole, defaultValue)
}

// GetSyntax returns the syntax value.
func (m metaWrapper) GetSyntax(defaultValue string) string {
	return m.original.GetDefault(domain.MetaKeySyntax, defaultValue)
}

// GetURL returns the URL value.
func (m metaWrapper) GetURL() string {
	if url, ok := m.original.Get(domain.MetaKeyURL); ok && len(url) > 0 {
		return url
	}
	return ""
}

// Pairs returns a list of all key/value pairs.
func (m metaWrapper) Pairs() []domain.MetaPair {
	return m.original.Pairs()
}

// PairsRest return a list of all key/value paris except the four basic ones.
func (m metaWrapper) PairsRest() []domain.MetaPair {
	return m.original.PairsRest()
}

// userWrapper is a wrapper around a user meta object.
type userWrapper struct {
	original *domain.Meta
}

func wrapUser(original *domain.Meta) userWrapper {
	return userWrapper{original}
}

// IsValid returns true, if user is a valid user
func (u userWrapper) IsValid() bool {
	return u.original != nil
}

// Zid returns the zettel ID of the wrapped user.
func (u userWrapper) Zid() domain.ZettelID {
	if orig := u.original; orig != nil {
		return orig.Zid
	}
	return domain.InvalidZettelID
}

// Ident returns the identifier (aka user name) of the user.
func (u userWrapper) Ident() string {
	return u.original.GetDefault(domain.MetaKeyIdent, "")
}

// IsOwner returns true, if the user is the owner of the zettelstore.
func (u userWrapper) IsOwner() bool {
	return u.IsValid() && u.Zid() == config.Owner()
}
