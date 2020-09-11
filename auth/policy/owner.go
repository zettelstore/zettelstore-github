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
