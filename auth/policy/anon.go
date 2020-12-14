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

type anonPolicy struct {
	expertMode    func() bool
	getVisibility func(*domain.Meta) domain.Visibility
	pre           Policy
}

func (ap *anonPolicy) CanReload(user *domain.Meta) bool {
	return ap.pre.CanReload(user)
}

func (ap *anonPolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	return ap.pre.CanCreate(user, newMeta)
}

func (ap *anonPolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	return ap.pre.CanRead(user, meta) &&
		(ap.getVisibility(meta) != domain.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	return ap.pre.CanWrite(user, oldMeta, newMeta) &&
		(ap.getVisibility(oldMeta) != domain.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return ap.pre.CanRename(user, meta) &&
		(ap.getVisibility(meta) != domain.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return ap.pre.CanDelete(user, meta) &&
		(ap.getVisibility(meta) != domain.VisibilityExpert || ap.expertMode())
}
