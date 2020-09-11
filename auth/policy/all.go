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

type allPolicy struct{}

func (a *allPolicy) CanReload(user *domain.Meta) bool {
	return true
}

func (a *allPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	return true
}

func (a *allPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	return true
}

func (a *allPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	return true
}

func (a *allPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return true
}

func (a *allPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return true
}
