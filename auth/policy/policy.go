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

// newPolicy creates a policy based on given constraints.
func newPolicy(
	withAuth func() bool,
	readonly bool,
	expertMode func() bool,
	isOwner func(domain.ZettelID) bool,
	getVisibility func(*domain.Meta) domain.Visibility,
) Policy {
	var pol Policy
	if readonly {
		pol = &roPolicy{}
	} else {
		pol = &defaultPolicy{}
	}
	if withAuth() {
		pol = &ownerPolicy{
			expertMode:    expertMode,
			isOwner:       isOwner,
			getVisibility: getVisibility,
			pre:           pol,
		}
	} else {
		pol = &anonPolicy{
			expertMode:    expertMode,
			getVisibility: getVisibility,
			pre:           pol,
		}
	}
	return &prePolicy{pol}
}

type prePolicy struct {
	post Policy
}

func (p *prePolicy) CanReload(user *domain.Meta) bool {
	return p.post.CanReload(user)
}

func (p *prePolicy) CanCreate(user *domain.Meta, newMeta *domain.Meta) bool {
	return newMeta != nil && p.post.CanCreate(user, newMeta)
}

func (p *prePolicy) CanRead(user *domain.Meta, meta *domain.Meta) bool {
	return meta != nil && p.post.CanRead(user, meta)
}

func (p *prePolicy) CanWrite(user *domain.Meta, oldMeta, newMeta *domain.Meta) bool {
	return oldMeta != nil && newMeta != nil && oldMeta.Zid == newMeta.Zid &&
		p.post.CanWrite(user, oldMeta, newMeta)
}

func (p *prePolicy) CanRename(user *domain.Meta, meta *domain.Meta) bool {
	return meta != nil && p.post.CanRename(user, meta)
}

func (p *prePolicy) CanDelete(user *domain.Meta, meta *domain.Meta) bool {
	return meta != nil && p.post.CanDelete(user, meta)
}
