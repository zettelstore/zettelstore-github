//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package runtime provides functions to retrieve runtime configuration data.
package runtime

import (
	"zettelstore.de/z/config/startup"
	"zettelstore.de/z/domain/meta"
)

var mapDefaultKeys = map[string]func() string{
	meta.MetaKeyCopyright: GetDefaultCopyright,
	meta.MetaKeyLang:      GetDefaultLang,
	meta.MetaKeyLicense:   GetDefaultLicense,
	meta.MetaKeyRole:      GetDefaultRole,
	meta.MetaKeySyntax:    GetDefaultSyntax,
	meta.MetaKeyTitle:     GetDefaultTitle,
}

// AddDefaultValues enriches the given meta data with its default values.
func AddDefaultValues(m *meta.Meta) *meta.Meta {
	result := m
	for k, f := range mapDefaultKeys {
		if _, ok := result.Get(k); !ok {
			if result == m {
				result = m.Clone()
			}
			if val := f(); len(val) > 0 || m.Type(k) == meta.MetaTypeEmpty {
				result.Set(k, val)
			}
		}
	}
	if result != m && m.IsFrozen() {
		result.Freeze()
	}
	return result
}

// GetTitle returns the value of the "title" key of the given meta. If there
// is no such value, GetDefaultTitle is returned.
func GetTitle(m *meta.Meta) string {
	if syntax, ok := m.Get(meta.MetaKeyTitle); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultTitle()
}

// GetRole returns the value of the "role" key of the given meta. If there
// is no such value, GetDefaultRole is returned.
func GetRole(m *meta.Meta) string {
	if syntax, ok := m.Get(meta.MetaKeyRole); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultRole()
}

// GetSyntax returns the value of the "syntax" key of the given meta. If there
// is no such value, GetDefaultSyntax is returned.
func GetSyntax(m *meta.Meta) string {
	if syntax, ok := m.Get(meta.MetaKeySyntax); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultSyntax()
}

// GetLang returns the value of the "lang" key of the given meta. If there is
// no such value, GetDefaultLang is returned.
func GetLang(m *meta.Meta) string {
	if lang, ok := m.Get(meta.MetaKeyLang); ok && len(lang) > 0 {
		return lang
	}
	return GetDefaultLang()
}

// GetVisibility returns the visibility value, or "login" if none is given.
func GetVisibility(m *meta.Meta) meta.Visibility {
	if val, ok := m.Get(meta.MetaKeyVisibility); ok {
		if vis := meta.GetVisibility(val); vis != meta.VisibilityUnknown {
			return vis
		}
	}
	return GetDefaultVisibility()
}

// GetUserRole role returns the user role of the given user zettel.
func GetUserRole(user *meta.Meta) meta.UserRole {
	if user == nil {
		if startup.WithAuth() {
			return meta.UserRoleUnknown
		}
		return meta.UserRoleOwner
	}
	if startup.IsOwner(user.Zid) {
		return meta.UserRoleOwner
	}
	if val, ok := user.Get(meta.MetaKeyUserRole); ok {
		if ur := meta.GetUserRole(val); ur != meta.UserRoleUnknown {
			return ur
		}
	}
	return meta.UserRoleReader
}
