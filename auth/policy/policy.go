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

// Policy is an interface for checking access authorization.
type Policy interface {
	// User is allowed to reload a place.
	CanReload(user *domain.Meta) bool

	// User is allowed to create a new zettel.
	CanCreate(user *domain.Meta, newMeta *domain.Meta) bool

	// User is allowed to read zettel
	CanRead(user *domain.Meta, meta *domain.Meta) bool

	// User is allowed to write zettel.
	CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool

	// User is allowed to rename zettel
	CanRename(user *domain.Meta, meta *domain.Meta) bool

	// User is allowed to delete zettel
	CanDelete(user *domain.Meta, meta *domain.Meta) bool
}

// NewPolicy creates a new policy object to check access autheorization.
func NewPolicy(name string) Policy {
	switch name {
	case "all":
		return &allPolicy{}
	}
	return &ownerPolicy{
		base:     &defaultPolicy{},
		owner:    config.Owner(),
		readonly: config.IsReadOnly(),
	}
}
