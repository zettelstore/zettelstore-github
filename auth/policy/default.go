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
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

type defaultPolicy struct{}

func (d *defaultPolicy) CanReload(user *domain.Meta) bool {
	return false
}

func (d *defaultPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	if user == nil || config.GetUserRole(user) == config.UserRoleReader {
		return false
	}
	if role, ok := newMeta.Get(domain.MetaKeyRole); ok && role == domain.MetaValueRoleUser {
		return false
	}
	return true
}

func (d *defaultPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
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

func (d *defaultPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	if !d.CanRead(user, oldMeta) {
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
	return d.CanCreate(user, newMeta)
}

func (d *defaultPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return false
}

func (d *defaultPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return false
}
