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

type ownerPolicy struct {
	owner domain.ZettelID
	pre   Policy
}

func (o *ownerPolicy) canDo(user *domain.Meta) bool {
	if o.owner.IsValid() {
		return user != nil && user.Zid == o.owner
	}
	return true
}

func (o *ownerPolicy) CanReload(user *domain.Meta) bool {
	if !o.pre.CanReload(user) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return false
}

func (o *ownerPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if newMeta == nil {
		return false
	}
	if !o.pre.CanCreate(user, newMeta) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) userCanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if user == nil || config.GetUserRole(user) == config.UserRoleReader {
		return false
	}
	if role, ok := newMeta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
		return false
	}
	return true
}

func (o *ownerPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	if meta == nil {
		return false
	}
	if !o.pre.CanRead(user, meta) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.userCanRead(user, meta)
}

func (o *ownerPolicy) userCanRead(user *domain.Meta, meta *domain.Meta) bool {
	switch visibility := config.GetVisibility(meta); visibility {
	case config.VisibilityOwner:
		return false
	case config.VisibilityPublic:
		return true
	}
	if user == nil {
		return false
	}
	role, ok := meta.Get(domain.MetaKeyRole)
	if !ok {
		return false
	}
	if role == domain.MetaValueRoleUser {
		// Only the user can read its own zettel
		return user.Zid == meta.Zid
	}
	return true
}

var noChangeUser = []string{
	domain.MetaKeyID,
	domain.MetaKeyRole,
	domain.MetaKeyUserID,
	domain.MetaKeyUserRole,
}

func (o *ownerPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	if oldMeta == nil || newMeta == nil || oldMeta.Zid != newMeta.Zid {
		return false
	}
	if !o.pre.CanWrite(user, oldMeta, newMeta) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	if !o.userCanRead(user, oldMeta) {
		return false
	}
	if user == nil {
		return false
	}
	if role, ok := oldMeta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
		if user.Zid != newMeta.Zid {
			return false
		}
		for _, key := range noChangeUser {
			if oldMeta.GetDefault(key, "") != newMeta.GetDefault(key, "") {
				return false
			}
		}
		return true
	}
	if config.GetUserRole(user) == config.UserRoleReader {
		return false
	}
	return o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	if meta == nil {
		return false
	}
	if !o.pre.CanRename(user, meta) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return false
}

func (o *ownerPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	if meta == nil {
		return false
	}
	if !o.pre.CanDelete(user, meta) {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return false
}
