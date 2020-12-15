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

// GetVisibility returns the visibility value, or "login" if none is given.
func GetVisibility(meta *domain.Meta) domain.Visibility {
	if val, ok := meta.Get(domain.MetaKeyVisibility); ok {
		if vis := domain.GetVisibility(val); vis != domain.VisibilityUnknown {
			return vis
		}
	}
	return GetDefaultVisibility()
}

// GetUserRole role returns the user role of the given user zettel.
func GetUserRole(user *domain.Meta) domain.UserRole {
	if user == nil {
		if startup.WithAuth() {
			return domain.UserRoleUnknown
		}
		return domain.UserRoleOwner
	}
	if startup.IsOwner(user.Zid) {
		return domain.UserRoleOwner
	}
	if val, ok := user.Get(domain.MetaKeyUserRole); ok {
		if ur := domain.GetUserRole(val); ur != domain.UserRoleUnknown {
			return ur
		}
	}
	return domain.UserRoleReader
}
