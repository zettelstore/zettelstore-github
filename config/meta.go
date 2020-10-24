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

// Package config provides functions to retrieve configuration data.
package config

import (
	"zettelstore.de/z/domain"
)

var mapDefaultKeys = map[string]func() string{
	domain.MetaKeyCopyright: GetDefaultCopyright,
	domain.MetaKeyLang:      GetDefaultLang,
	domain.MetaKeyLicense:   GetDefaultLicense,
	domain.MetaKeyRole:      GetDefaultRole,
	domain.MetaKeySyntax:    GetDefaultSyntax,
	domain.MetaKeyTitle:     GetDefaultTitle,
}

// AddDefaultValues enriches the given meta data with its default values.
func AddDefaultValues(meta *domain.Meta) *domain.Meta {
	result := meta
	for k, f := range mapDefaultKeys {
		if _, ok := result.Get(k); !ok {
			if result == meta {
				result = meta.Clone()
			}
			if val := f(); len(val) > 0 || meta.Type(k) == domain.MetaTypeEmpty {
				result.Set(k, val)
			}
		}
	}
	if result != meta && meta.IsFrozen() {
		result.Freeze()
	}
	return result
}

// GetTitle returns the value of the "title" key of the given meta. If there
// is no such value, GetDefaultTitle is returned.
func GetTitle(meta *domain.Meta) string {
	if syntax, ok := meta.Get(domain.MetaKeyTitle); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultTitle()
}

// GetRole returns the value of the "role" key of the given meta. If there
// is no such value, GetDefaultRole is returned.
func GetRole(meta *domain.Meta) string {
	if syntax, ok := meta.Get(domain.MetaKeyRole); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultRole()
}

// GetSyntax returns the value of the "syntax" key of the given meta. If there
// is no such value, GetDefaultSyntax is returned.
func GetSyntax(meta *domain.Meta) string {
	if syntax, ok := meta.Get(domain.MetaKeySyntax); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultSyntax()
}

// GetLang returns the value of the "lang" key of the given meta. If there is
// no such value, GetDefaultLang is returned.
func GetLang(meta *domain.Meta) string {
	if lang, ok := meta.Get(domain.MetaKeyLang); ok && len(lang) > 0 {
		return lang
	}
	return GetDefaultLang()
}

// Visibility enumerates the variations of the 'visibility' meta key.
type Visibility int

// Supported values for visibility.
const (
	_ Visibility = iota
	VisibilityPublic
	VisibilityLogin
	VisibilityOwner
)

var visMap = map[string]Visibility{
	domain.MetaValueVisibilityPublic: VisibilityPublic,
	domain.MetaValueVisibilityLogin:  VisibilityLogin,
	domain.MetaValueVisibilityOwner:  VisibilityOwner,
}

// GetVisibility returns the visibility value, or "login" if none is given.
func GetVisibility(meta *domain.Meta) Visibility {
	if val, ok := meta.Get(domain.MetaKeyVisibility); ok {
		if vis, ok := visMap[val]; ok {
			return vis
		}
	}
	return VisibilityLogin
}

// UserRole enumerates the supported values of meta key 'user-role'.
type UserRole int

// Supported values for user roles.
const (
	_ UserRole = iota
	UserRoleReader
	UserRoleWriter
)

var urMap = map[string]UserRole{
	"reader": UserRoleReader,
	"writer": UserRoleWriter,
}

// GetUserRole role returns the user role of the given user zettel.
func GetUserRole(user *domain.Meta) UserRole {
	if val, ok := user.Get(domain.MetaKeyUserRole); ok {
		if ur, ok := urMap[val]; ok {
			return ur
		}
	}
	return UserRoleReader
}
