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
	"fmt"
	"testing"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

func TestPolicies(t *testing.T) {
	testScene := []struct {
		name     string
		readonly bool
		withAuth bool
		pol      Policy
	}{
		{"read-only/no auth", true, false, newPolicy(withoutAuth, true, nil, nil)},
		{"read-only/with auth", true, true, newPolicy(withAuth, true, isOwner, getVisibility)},
		{"write/no auth", false, false, newPolicy(withoutAuth, false, nil, nil)},
		{"write/with auth", false, true, newPolicy(withAuth, false, isOwner, getVisibility)},
	}
	for _, ts := range testScene {
		t.Run(ts.name, func(tt *testing.T) {
			testReload(tt, ts.pol, ts.withAuth, ts.readonly)
			testCreate(tt, ts.pol, ts.withAuth, ts.readonly)
			testRead(tt, ts.pol, ts.withAuth, ts.readonly)
			testWrite(tt, ts.pol, ts.withAuth, ts.readonly)
			testDelete(tt, ts.pol, ts.withAuth, ts.readonly)
			testRename(tt, ts.pol, ts.withAuth, ts.readonly)
		})
	}
}

func withAuth() bool                   { return true }
func withoutAuth() bool                { return false }
func isOwner(zid domain.ZettelID) bool { return zid == ownerZid }
func getVisibility(meta *domain.Meta) config.Visibility {
	if vis, ok := meta.Get(domain.MetaKeyVisibility); ok {
		switch vis {
		case "public":
			return config.VisibilityPublic
		case "owner":
			return config.VisibilityOwner
		}
	}
	return config.VisibilityLogin
}

func testReload(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	testCases := []struct {
		user *domain.Meta
		exp  bool
	}{
		{newAnon(), !withAuth},
		{newReader(), !withAuth},
		{newWriter(), !withAuth},
		{newOwner(), true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Reload/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			got := pol.CanReload(tc.user)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testCreate(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	userZettel := newUserZettel()
	testCases := []struct {
		user *domain.Meta
		meta *domain.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false}, {reader, nil, false}, {writer, nil, false}, {owner, nil, false},
		// Ordinary zettel
		{anonUser, zettel, !withAuth && !readonly},
		{reader, zettel, !withAuth && !readonly},
		{writer, zettel, !readonly},
		{owner, zettel, !readonly},
		// User zettel
		{anonUser, userZettel, !withAuth && !readonly},
		{reader, userZettel, !withAuth && !readonly},
		{writer, userZettel, !withAuth && !readonly},
		{owner, userZettel, !readonly},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Create/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			got := pol.CanCreate(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testRead(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	publicZettel := newPublicZettel()
	loginZettel := newLoginZettel()
	ownerZettel := newOwnerZettel()
	userZettel := newUserZettel()
	testCases := []struct {
		user *domain.Meta
		meta *domain.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false}, {reader, nil, false}, {writer, nil, false}, {owner, nil, false},
		// Ordinary zettel
		{anonUser, zettel, !withAuth},
		{reader, zettel, true},
		{writer, zettel, true},
		{owner, zettel, true},
		// Public zettel
		{anonUser, publicZettel, true},
		{reader, publicZettel, true},
		{writer, publicZettel, true},
		{owner, publicZettel, true},
		// Login zettel
		{anonUser, loginZettel, !withAuth},
		{reader, loginZettel, true},
		{writer, loginZettel, true},
		{owner, loginZettel, true},
		// Owner zettel
		{anonUser, ownerZettel, !withAuth},
		{reader, ownerZettel, !withAuth},
		{writer, ownerZettel, !withAuth},
		{owner, ownerZettel, true},
		// Other user zettel
		{anonUser, userZettel, !withAuth},
		{reader, userZettel, !withAuth},
		{writer, userZettel, !withAuth},
		{owner, userZettel, true},
		// Own user zettel
		{reader, reader, true},
		{writer, writer, true},
		{owner, owner, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Read/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			got := pol.CanRead(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testWrite(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	publicZettel := newPublicZettel()
	loginZettel := newLoginZettel()
	ownerZettel := newOwnerZettel()
	userZettel := newUserZettel()
	writerNew := writer.Clone()
	writerNew.Set(domain.MetaKeyUserRole, owner.GetDefault(domain.MetaKeyUserRole, ""))
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *domain.Meta
		old  *domain.Meta
		new  *domain.Meta
		exp  bool
	}{
		// No old and new meta
		{anonUser, nil, nil, false}, {reader, nil, nil, false}, {writer, nil, nil, false}, {owner, nil, nil, false},
		// No old meta
		{anonUser, nil, zettel, false}, {reader, nil, zettel, false}, {writer, nil, zettel, false}, {owner, nil, zettel, false},
		// No new meta
		{anonUser, zettel, nil, false}, {reader, zettel, nil, false}, {writer, zettel, nil, false}, {owner, zettel, nil, false},
		// Old an new zettel have different zettel identifier
		{anonUser, zettel, publicZettel, false},
		{reader, zettel, publicZettel, false},
		{writer, zettel, publicZettel, false},
		{owner, zettel, publicZettel, false},
		// Overwrite a normal zettel
		{anonUser, zettel, zettel, !withAuth && !readonly},
		{reader, zettel, zettel, !withAuth && !readonly},
		{writer, zettel, zettel, !readonly},
		{owner, zettel, zettel, !readonly},
		// Public zettel
		{anonUser, publicZettel, publicZettel, !withAuth && !readonly},
		{reader, publicZettel, publicZettel, !withAuth && !readonly},
		{writer, publicZettel, publicZettel, !readonly},
		{owner, publicZettel, publicZettel, !readonly},
		// Login zettel
		{anonUser, loginZettel, loginZettel, !withAuth && !readonly},
		{reader, loginZettel, loginZettel, !withAuth && !readonly},
		{writer, loginZettel, loginZettel, !readonly},
		{owner, loginZettel, loginZettel, !readonly},
		// Owner zettel
		{anonUser, ownerZettel, ownerZettel, !withAuth && !readonly},
		{reader, ownerZettel, ownerZettel, !withAuth && !readonly},
		{writer, ownerZettel, ownerZettel, !withAuth && !readonly},
		{owner, ownerZettel, ownerZettel, !readonly},
		// Other user zettel
		{anonUser, userZettel, userZettel, !withAuth && !readonly},
		{reader, userZettel, userZettel, !withAuth && !readonly},
		{writer, userZettel, userZettel, !withAuth && !readonly},
		{owner, userZettel, userZettel, !readonly},
		// Own user zettel
		{reader, reader, reader, !readonly},
		{writer, writer, writer, !readonly},
		{owner, owner, owner, !readonly},
		// Writer cannot change importand metadata of its own user zettel
		{writer, writer, writerNew, !withAuth && !readonly},
		// No r/o zettel
		{anonUser, roFalse, roFalse, !withAuth && !readonly},
		{reader, roFalse, roFalse, !withAuth && !readonly},
		{writer, roFalse, roFalse, !readonly},
		{owner, roFalse, roFalse, !readonly},
		// Reader r/o zettel
		{anonUser, roReader, roReader, false},
		{reader, roReader, roReader, false},
		{writer, roReader, roReader, !readonly},
		{owner, roReader, roReader, !readonly},
		// Writer r/o zettel
		{anonUser, roWriter, roWriter, false},
		{reader, roWriter, roWriter, false},
		{writer, roWriter, roWriter, false},
		{owner, roWriter, roWriter, !readonly},
		// Owner r/o zettel
		{anonUser, roOwner, roOwner, false},
		{reader, roOwner, roOwner, false},
		{writer, roOwner, roOwner, false},
		{owner, roOwner, roOwner, false},
		// r/o = true zettel
		{anonUser, roTrue, roTrue, false},
		{reader, roTrue, roTrue, false},
		{writer, roTrue, roTrue, false},
		{owner, roTrue, roTrue, false},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Write/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			if i == 34 {
				i = 0
			}
			got := pol.CanWrite(tc.user, tc.old, tc.new)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testDelete(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *domain.Meta
		meta *domain.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false}, {reader, nil, false}, {writer, nil, false}, {owner, nil, false},
		// Any zettel
		{anonUser, zettel, !withAuth && !readonly},
		{reader, zettel, !withAuth && !readonly},
		{writer, zettel, !withAuth && !readonly},
		{owner, zettel, !readonly},
		// No r/o zettel
		{anonUser, roFalse, !withAuth && !readonly},
		{reader, roFalse, !withAuth && !readonly},
		{writer, roFalse, !withAuth && !readonly},
		{owner, roFalse, !readonly},
		// Reader r/o zettel
		{anonUser, roReader, false},
		{reader, roReader, false},
		{writer, roReader, !withAuth && !readonly},
		{owner, roReader, !readonly},
		// Writer r/o zettel
		{anonUser, roWriter, false}, {reader, roWriter, false}, {writer, roWriter, false}, {owner, roWriter, !readonly},
		// Owner r/o zettel
		{anonUser, roOwner, false}, {reader, roOwner, false}, {writer, roOwner, false}, {owner, roOwner, false},
		// r/o = true zettel
		{anonUser, roTrue, false}, {reader, roTrue, false}, {writer, roTrue, false}, {owner, roTrue, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Delete/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			got := pol.CanDelete(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testRename(t *testing.T, pol Policy, withAuth bool, readonly bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *domain.Meta
		meta *domain.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false}, {reader, nil, false}, {writer, nil, false}, {owner, nil, false},
		// Any zettel
		{anonUser, zettel, !withAuth && !readonly},
		{reader, zettel, !withAuth && !readonly},
		{writer, zettel, !withAuth && !readonly},
		{owner, zettel, !readonly},
		// No r/o zettel
		{anonUser, roFalse, !withAuth && !readonly},
		{reader, roFalse, !withAuth && !readonly},
		{writer, roFalse, !withAuth && !readonly},
		{owner, roFalse, !readonly},
		// Reader r/o zettel
		{anonUser, roReader, false},
		{reader, roReader, false},
		{writer, roReader, !withAuth && !readonly},
		{owner, roReader, !readonly},
		// Writer r/o zettel
		{anonUser, roWriter, false}, {reader, roWriter, false}, {writer, roWriter, false}, {owner, roWriter, !readonly},
		// Owner r/o zettel
		{anonUser, roOwner, false}, {reader, roOwner, false}, {writer, roOwner, false}, {owner, roOwner, false},
		// r/o = true zettel
		{anonUser, roTrue, false}, {reader, roTrue, false}, {writer, roTrue, false}, {owner, roTrue, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Rename/auth=%v/readonly=%v", withAuth, readonly), func(tt *testing.T) {
			got := pol.CanRename(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

const (
	readerZid = domain.ZettelID(1013)
	writerZid = domain.ZettelID(1015)
	ownerZid  = domain.ZettelID(1017)
	zettelZid = domain.ZettelID(1021)
	visZid    = domain.ZettelID(1023)
	userZid   = domain.ZettelID(1025)
)

func newAnon() *domain.Meta { return nil }
func newReader() *domain.Meta {
	user := domain.NewMeta(readerZid)
	user.Set(domain.MetaKeyTitle, "Reader")
	user.Set(domain.MetaKeyRole, domain.MetaValueRoleUser)
	user.Set(domain.MetaKeyUserRole, "reader")
	return user
}
func newWriter() *domain.Meta {
	user := domain.NewMeta(writerZid)
	user.Set(domain.MetaKeyTitle, "Writer")
	user.Set(domain.MetaKeyRole, domain.MetaValueRoleUser)
	user.Set(domain.MetaKeyUserRole, "writer")
	return user
}
func newOwner() *domain.Meta {
	user := domain.NewMeta(ownerZid)
	user.Set(domain.MetaKeyTitle, "Owner")
	user.Set(domain.MetaKeyRole, domain.MetaValueRoleUser)
	user.Set(domain.MetaKeyUserRole, "owner")
	return user
}
func newZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "Any Zettel")
	return meta
}
func newPublicZettel() *domain.Meta {
	meta := domain.NewMeta(visZid)
	meta.Set(domain.MetaKeyTitle, "Public Zettel")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityPublic)
	return meta
}
func newLoginZettel() *domain.Meta {
	meta := domain.NewMeta(visZid)
	meta.Set(domain.MetaKeyTitle, "Login Zettel")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityLogin)
	return meta
}
func newOwnerZettel() *domain.Meta {
	meta := domain.NewMeta(visZid)
	meta.Set(domain.MetaKeyTitle, "Owner Zettel")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityOwner)
	return meta
}
func newRoFalseZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "No r/o Zettel")
	meta.Set(domain.MetaKeyReadOnly, "false")
	return meta
}
func newRoTrueZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "A r/o Zettel")
	meta.Set(domain.MetaKeyReadOnly, "true")
	return meta
}
func newRoReaderZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "Reader r/o Zettel")
	meta.Set(domain.MetaKeyReadOnly, "reader")
	return meta
}
func newRoWriterZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "Writer r/o Zettel")
	meta.Set(domain.MetaKeyReadOnly, "writer")
	return meta
}
func newRoOwnerZettel() *domain.Meta {
	meta := domain.NewMeta(zettelZid)
	meta.Set(domain.MetaKeyTitle, "Owner r/o Zettel")
	meta.Set(domain.MetaKeyReadOnly, "owner")
	return meta
}
func newUserZettel() *domain.Meta {
	meta := domain.NewMeta(userZid)
	meta.Set(domain.MetaKeyTitle, "Any User")
	meta.Set(domain.MetaKeyRole, domain.MetaValueRoleUser)
	return meta
}
