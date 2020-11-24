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
