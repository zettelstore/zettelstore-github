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
	expertMode    func() bool
	isOwner       func(domain.ZettelID) bool
	getVisibility func(*domain.Meta) domain.Visibility
	pre           Policy
}

func (o *ownerPolicy) CanReload(user *domain.Meta) bool {
	// No need to call o.pre.CanReload(user), because it will always return true.
	// Both the default and the readonly policy allow to reload a place.

	// Only the owner is allowed to reload a place
	return user != nil && o.isOwner(user.Zid)
}

func (o *ownerPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if user == nil || !o.pre.CanCreate(user, newMeta) {
		return false
	}
	return o.isOwner(user.Zid) || o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) userCanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if config.GetUserRole(user) == domain.UserRoleReader {
		return false
	}
	if role, ok := newMeta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
		return false
	}
	return true
}

func (o *ownerPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	// No need to call o.pre.CanRead(user, meta), because it will always return true.
	// Both the default and the readonly policy allow to read a zettel.
	vis := o.getVisibility(meta)
	if vis == domain.VisibilityExpert && !o.expertMode() {
		return false
	}
	return (user != nil && o.isOwner(user.Zid)) || o.userCanRead(user, meta, vis)
}

func (o *ownerPolicy) userCanRead(
	user *domain.Meta, meta *domain.Meta, vis domain.Visibility) bool {
	switch vis {
	case domain.VisibilityOwner, domain.VisibilityExpert:
		return false
	case domain.VisibilityPublic:
		return true
	}
	if user == nil {
		return false
	}
	if role, ok := meta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
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
	if user == nil || !o.pre.CanWrite(user, oldMeta, newMeta) {
		return false
	}
	vis := o.getVisibility(oldMeta)
	if vis == domain.VisibilityExpert && !o.expertMode() {
		return false
	}
	if o.isOwner(user.Zid) {
		return true
	}
	if !o.userCanRead(user, oldMeta, vis) {
		return false
	}
	if role, ok := oldMeta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
		// Here we know, that user.Zid == newMeta.Zid (because of userCanRead) and
		// user.Zid == newMeta.Zid (because oldMeta.Zid == newMeta.Zid)
		for _, key := range noChangeUser {
			if oldMeta.GetDefault(key, "") != newMeta.GetDefault(key, "") {
				return false
			}
		}
		return true
	}
	if config.GetUserRole(user) == domain.UserRoleReader {
		return false
	}
	return o.userCanCreate(user, newMeta)
}

func (o *ownerPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	if user == nil || !o.pre.CanRename(user, meta) {
		return false
	}
	if o.getVisibility(meta) == domain.VisibilityExpert && !o.expertMode() {
		return false
	}
	return o.isOwner(user.Zid)
}

func (o *ownerPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	if user == nil || !o.pre.CanDelete(user, meta) {
		return false
	}
	if o.getVisibility(meta) == domain.VisibilityExpert && !o.expertMode() {
		return false
	}
	return o.isOwner(user.Zid)
}
