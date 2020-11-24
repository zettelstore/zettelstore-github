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
	"zettelstore.de/z/domain"
)

type ownerPolicy struct {
	base     Policy
	owner    domain.ZettelID
	readonly bool
}

func (o *ownerPolicy) canDo(user *domain.Meta) bool {
	if o.owner.IsValid() {
		return user != nil && user.Zid == o.owner
	}
	return true
}

func (o *ownerPolicy) CanReload(user *domain.Meta) bool {
	if o.canDo(user) {
		return true
	}
	return o.base.CanReload(user)
}

func (o *ownerPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if o.readonly || newMeta == nil {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.base.CanCreate(user, newMeta)
}

func (o *ownerPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	if meta == nil {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.base.CanRead(user, meta)
}

func (o *ownerPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	if o.readonly || oldMeta == nil || newMeta == nil || oldMeta.Zid != newMeta.Zid {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.base.CanWrite(user, oldMeta, newMeta)
}

func (o *ownerPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	if o.readonly || meta == nil {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.base.CanRename(user, meta)
}

func (o *ownerPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	if o.readonly || meta == nil {
		return false
	}
	if o.canDo(user) {
		return true
	}
	return o.base.CanDelete(user, meta)
}
