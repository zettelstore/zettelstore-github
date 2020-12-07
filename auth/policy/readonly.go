//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package policy provides some interfaces and implementation for authorization policies.
package policy

import (
	"zettelstore.de/z/domain"
)

type roPolicy struct{}

func (p *roPolicy) CanReload(user *domain.Meta) bool {
	return true
}

func (p *roPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	return false
}

func (p *roPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	return true
}

func (p *roPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	return false
}

func (p *roPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return false
}

func (p *roPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return false
}
