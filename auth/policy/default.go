//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package policy provides some interfaces and implementation for authorizsation policies.
package policy

import (
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

type defaultPolicy struct{}

func (d *defaultPolicy) CanReload(user *domain.Meta) bool {
	return true
}

func (d *defaultPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	return true
}

func (d *defaultPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	return true
}

func (d *defaultPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	return d.canChange(user, oldMeta)
}

func (d *defaultPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return d.canChange(user, meta)
}

func (d *defaultPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return d.canChange(user, meta)
}

func (d *defaultPolicy) canChange(user *domain.Meta, meta *domain.Meta) bool {
	metaRo, ok := meta.Get(domain.MetaKeyReadOnly)
	if !ok {
		return true
	}
	if user == nil {
		// If we are here, there is no authentication.
		// See owner.go:CanWrite.

		// No authentication: check for owner-like restriction, since the user acts as an owner
		if metaRo == "owner" || domain.BoolValue(metaRo) {
			return false
		}
		return true
	}

	userRole := config.GetUserRole(user)
	switch metaRo {
	case "reader":
		return userRole > config.UserRoleReader
	case "writer":
		return userRole > config.UserRoleWriter
	case "owner":
		return userRole > config.UserRoleOwner
	}
	return !domain.BoolValue(metaRo)
}
