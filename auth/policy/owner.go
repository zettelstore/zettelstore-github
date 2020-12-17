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
	"zettelstore.de/z/config/runtime"
	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
)

type ownerPolicy struct {
	expertMode    func() bool
	isOwner       func(id.ZettelID) bool
	getVisibility func(*meta.Meta) meta.Visibility
	pre           Policy
}

func (o *ownerPolicy) CanReload(user *meta.Meta) bool {
	// No need to call o.pre.CanReload(user), because it will always return true.
	// Both the default and the readonly policy allow to reload a place.

	// Only the owner is allowed to reload a place
	return user != nil && o.isOwner(user.Zid)
}

func (o *ownerPolicy) CanCreate(user *meta.Meta, newMeta *meta.Meta) bool {
	if user == nil || !o.pre.CanCreate(user, newMeta) {
		return false
	}
	return o.isOwner(user.Zid) || o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) userCanCreate(user *meta.Meta, newMeta *meta.Meta) bool {
	if runtime.GetUserRole(user) == meta.UserRoleReader {
		return false
	}
	if role, ok := newMeta.Get(meta.MetaKeyRole); ok && role == meta.MetaValueRoleUser {
		return false
	}
	return true
}

func (o *ownerPolicy) CanRead(user *meta.Meta, m *meta.Meta) bool {
	// No need to call o.pre.CanRead(user, meta), because it will always return true.
	// Both the default and the readonly policy allow to read a zettel.
	vis := o.getVisibility(m)
	if vis == meta.VisibilityExpert && !o.expertMode() {
		return false
	}
	return (user != nil && o.isOwner(user.Zid)) || o.userCanRead(user, m, vis)
}

func (o *ownerPolicy) userCanRead(user *meta.Meta, m *meta.Meta, vis meta.Visibility) bool {
	switch vis {
	case meta.VisibilityOwner, meta.VisibilityExpert:
		return false
	case meta.VisibilityPublic:
		return true
	}
	if user == nil {
		return false
	}
	if role, ok := m.Get(meta.MetaKeyRole); ok && role == meta.MetaValueRoleUser {
		// Only the user can read its own zettel
		return user.Zid == m.Zid
	}
	return true
}

var noChangeUser = []string{
	meta.MetaKeyID,
	meta.MetaKeyRole,
	meta.MetaKeyUserID,
	meta.MetaKeyUserRole,
}

func (o *ownerPolicy) CanWrite(user *meta.Meta, oldMeta, newMeta *meta.Meta) bool {
	if user == nil || !o.pre.CanWrite(user, oldMeta, newMeta) {
		return false
	}
	vis := o.getVisibility(oldMeta)
	if vis == meta.VisibilityExpert && !o.expertMode() {
		return false
	}
	if o.isOwner(user.Zid) {
		return true
	}
	if !o.userCanRead(user, oldMeta, vis) {
		return false
	}
	if role, ok := oldMeta.Get(meta.MetaKeyRole); ok && role == meta.MetaValueRoleUser {
		// Here we know, that user.Zid == newMeta.Zid (because of userCanRead) and
		// user.Zid == newMeta.Zid (because oldMeta.Zid == newMeta.Zid)
		for _, key := range noChangeUser {
			if oldMeta.GetDefault(key, "") != newMeta.GetDefault(key, "") {
				return false
			}
		}
		return true
	}
	if runtime.GetUserRole(user) == meta.UserRoleReader {
		return false
	}
	return o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) CanRename(user *meta.Meta, m *meta.Meta) bool {
	if user == nil || !o.pre.CanRename(user, m) {
		return false
	}
	if o.getVisibility(m) == meta.VisibilityExpert && !o.expertMode() {
		return false
	}
	return o.isOwner(user.Zid)
}

func (o *ownerPolicy) CanDelete(user *meta.Meta, m *meta.Meta) bool {
	if user == nil || !o.pre.CanDelete(user, m) {
		return false
	}
	if o.getVisibility(m) == meta.VisibilityExpert && !o.expertMode() {
		return false
	}
	return o.isOwner(user.Zid)
}
