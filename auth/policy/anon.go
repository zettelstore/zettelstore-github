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
	"zettelstore.de/z/domain/meta"
)

type anonPolicy struct {
	expertMode    func() bool
	getVisibility func(*meta.Meta) meta.Visibility
	pre           Policy
}

func (ap *anonPolicy) CanReload(user *meta.Meta) bool {
	return ap.pre.CanReload(user)
}

func (ap *anonPolicy) CanCreate(user *meta.Meta, newMeta *meta.Meta) bool {
	return ap.pre.CanCreate(user, newMeta)
}

func (ap *anonPolicy) CanRead(user *meta.Meta, m *meta.Meta) bool {
	return ap.pre.CanRead(user, m) &&
		(ap.getVisibility(m) != meta.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanWrite(user *meta.Meta, oldMeta, newMeta *meta.Meta) bool {
	return ap.pre.CanWrite(user, oldMeta, newMeta) &&
		(ap.getVisibility(oldMeta) != meta.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanRename(user *meta.Meta, m *meta.Meta) bool {
	return ap.pre.CanRename(user, m) &&
		(ap.getVisibility(m) != meta.VisibilityExpert || ap.expertMode())
}

func (ap *anonPolicy) CanDelete(user *meta.Meta, m *meta.Meta) bool {
	return ap.pre.CanDelete(user, m) &&
		(ap.getVisibility(m) != meta.VisibilityExpert || ap.expertMode())
}
