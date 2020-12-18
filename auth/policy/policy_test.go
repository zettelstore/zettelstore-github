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

	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
)

func TestPolicies(t *testing.T) {
	testScene := []struct {
		name     string
		readonly bool
		withAuth bool
		isExpert bool
		pol      Policy
	}{
		{"read-only/no auth", true, false, false, newPolicy(
			withoutAuth, true, noExpertMode, isOwner, getVisibility)},
		{"read-only/no auth/expert", true, false, true, newPolicy(
			withoutAuth, true, expertMode, isOwner, getVisibility)},
		{"read-only/with auth", true, true, false, newPolicy(
			withAuth, true, noExpertMode, isOwner, getVisibility)},
		{"read-only/with auth/expert", true, true, true, newPolicy(
			withAuth, true, expertMode, isOwner, getVisibility)},
		{"write/no auth", false, false, false, newPolicy(
			withoutAuth, false, noExpertMode, isOwner, getVisibility)},
		{"write/no auth/expert", false, false, true, newPolicy(
			withoutAuth, false, expertMode, isOwner, getVisibility)},
		{"write/with auth", false, true, false, newPolicy(
			withAuth, false, noExpertMode, isOwner, getVisibility)},
		{"write/with auth/expert", false, true, true, newPolicy(
			withAuth, false, expertMode, isOwner, getVisibility)},
	}
	for _, ts := range testScene {
		t.Run(ts.name, func(tt *testing.T) {
			testReload(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
			testCreate(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
			testRead(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
			testWrite(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
			testRename(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
			testDelete(tt, ts.pol, ts.withAuth, ts.readonly, ts.isExpert)
		})
	}
}

func withAuth() bool          { return true }
func withoutAuth() bool       { return false }
func expertMode() bool        { return true }
func noExpertMode() bool      { return false }
func isOwner(zid id.Zid) bool { return zid == ownerZid }
func getVisibility(m *meta.Meta) meta.Visibility {
	if vis, ok := m.Get(meta.KeyVisibility); ok {
		switch vis {
		case meta.ValueVisibilityPublic:
			return meta.VisibilityPublic
		case meta.ValueVisibilityOwner:
			return meta.VisibilityOwner
		case meta.ValueVisibilityExpert:
			return meta.VisibilityExpert
		}
	}
	return meta.VisibilityLogin
}

func testReload(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	testCases := []struct {
		user *meta.Meta
		exp  bool
	}{
		{newAnon(), !withAuth},
		{newReader(), !withAuth},
		{newWriter(), !withAuth},
		{newOwner(), true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Reload/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
			got := pol.CanReload(tc.user)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testCreate(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	userZettel := newUserZettel()
	testCases := []struct {
		user *meta.Meta
		meta *meta.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false},
		{reader, nil, false},
		{writer, nil, false},
		{owner, nil, false},
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
		t.Run(fmt.Sprintf("Create/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
			got := pol.CanCreate(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testRead(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	publicZettel := newPublicZettel()
	loginZettel := newLoginZettel()
	ownerZettel := newOwnerZettel()
	expertZettel := newExpertZettel()
	userZettel := newUserZettel()
	testCases := []struct {
		user *meta.Meta
		meta *meta.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false},
		{reader, nil, false},
		{writer, nil, false},
		{owner, nil, false},
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
		// // Expert zettel
		{anonUser, expertZettel, !withAuth && isExpert},
		{reader, expertZettel, !withAuth && isExpert},
		{writer, expertZettel, !withAuth && isExpert},
		{owner, expertZettel, isExpert},
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
		t.Run(fmt.Sprintf("Read/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
			got := pol.CanRead(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testWrite(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	publicZettel := newPublicZettel()
	loginZettel := newLoginZettel()
	ownerZettel := newOwnerZettel()
	expertZettel := newExpertZettel()
	userZettel := newUserZettel()
	writerNew := writer.Clone()
	writerNew.Set(meta.KeyUserRole, owner.GetDefault(meta.KeyUserRole, ""))
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *meta.Meta
		old  *meta.Meta
		new  *meta.Meta
		exp  bool
	}{
		// No old and new meta
		{anonUser, nil, nil, false},
		{reader, nil, nil, false},
		{writer, nil, nil, false},
		{owner, nil, nil, false},
		// No old meta
		{anonUser, nil, zettel, false},
		{reader, nil, zettel, false},
		{writer, nil, zettel, false},
		{owner, nil, zettel, false},
		// No new meta
		{anonUser, zettel, nil, false},
		{reader, zettel, nil, false},
		{writer, zettel, nil, false},
		{owner, zettel, nil, false},
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
		// // Expert zettel
		{anonUser, expertZettel, expertZettel, !withAuth && !readonly && isExpert},
		{reader, expertZettel, expertZettel, !withAuth && !readonly && isExpert},
		{writer, expertZettel, expertZettel, !withAuth && !readonly && isExpert},
		{owner, expertZettel, expertZettel, !readonly && isExpert},
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
		t.Run(fmt.Sprintf("Write/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
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

func testRename(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	expertZettel := newExpertZettel()
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *meta.Meta
		meta *meta.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false},
		{reader, nil, false},
		{writer, nil, false},
		{owner, nil, false},
		// Any zettel
		{anonUser, zettel, !withAuth && !readonly},
		{reader, zettel, !withAuth && !readonly},
		{writer, zettel, !withAuth && !readonly},
		{owner, zettel, !readonly},
		// Expert zettel
		{anonUser, expertZettel, !withAuth && !readonly && isExpert},
		{reader, expertZettel, !withAuth && !readonly && isExpert},
		{writer, expertZettel, !withAuth && !readonly && isExpert},
		{owner, expertZettel, !readonly && isExpert},
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
		{anonUser, roWriter, false},
		{reader, roWriter, false},
		{writer, roWriter, false},
		{owner, roWriter, !readonly},
		// Owner r/o zettel
		{anonUser, roOwner, false},
		{reader, roOwner, false},
		{writer, roOwner, false},
		{owner, roOwner, false},
		// r/o = true zettel
		{anonUser, roTrue, false},
		{reader, roTrue, false},
		{writer, roTrue, false},
		{owner, roTrue, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Rename/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
			got := pol.CanRename(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

func testDelete(t *testing.T, pol Policy, withAuth bool, readonly bool, isExpert bool) {
	t.Helper()
	anonUser := newAnon()
	reader := newReader()
	writer := newWriter()
	owner := newOwner()
	zettel := newZettel()
	expertZettel := newExpertZettel()
	roFalse := newRoFalseZettel()
	roTrue := newRoTrueZettel()
	roReader := newRoReaderZettel()
	roWriter := newRoWriterZettel()
	roOwner := newRoOwnerZettel()
	testCases := []struct {
		user *meta.Meta
		meta *meta.Meta
		exp  bool
	}{
		// No meta
		{anonUser, nil, false},
		{reader, nil, false},
		{writer, nil, false},
		{owner, nil, false},
		// Any zettel
		{anonUser, zettel, !withAuth && !readonly},
		{reader, zettel, !withAuth && !readonly},
		{writer, zettel, !withAuth && !readonly},
		{owner, zettel, !readonly},
		// Expert zettel
		{anonUser, expertZettel, !withAuth && !readonly && isExpert},
		{reader, expertZettel, !withAuth && !readonly && isExpert},
		{writer, expertZettel, !withAuth && !readonly && isExpert},
		{owner, expertZettel, !readonly && isExpert},
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
		{anonUser, roWriter, false},
		{reader, roWriter, false},
		{writer, roWriter, false},
		{owner, roWriter, !readonly},
		// Owner r/o zettel
		{anonUser, roOwner, false},
		{reader, roOwner, false},
		{writer, roOwner, false},
		{owner, roOwner, false},
		// r/o = true zettel
		{anonUser, roTrue, false},
		{reader, roTrue, false},
		{writer, roTrue, false},
		{owner, roTrue, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Delete/auth=%v/readonly=%v/expert=%v",
			withAuth, readonly, isExpert), func(tt *testing.T) {
			got := pol.CanDelete(tc.user, tc.meta)
			if tc.exp != got {
				tt.Errorf("exp=%v, but got=%v", tc.exp, got)
			}
		})
	}
}

const (
	readerZid = id.Zid(1013)
	writerZid = id.Zid(1015)
	ownerZid  = id.Zid(1017)
	zettelZid = id.Zid(1021)
	visZid    = id.Zid(1023)
	userZid   = id.Zid(1025)
)

func newAnon() *meta.Meta { return nil }
func newReader() *meta.Meta {
	user := meta.NewMeta(readerZid)
	user.Set(meta.KeyTitle, "Reader")
	user.Set(meta.KeyRole, meta.ValueRoleUser)
	user.Set(meta.KeyUserRole, "reader")
	return user
}
func newWriter() *meta.Meta {
	user := meta.NewMeta(writerZid)
	user.Set(meta.KeyTitle, "Writer")
	user.Set(meta.KeyRole, meta.ValueRoleUser)
	user.Set(meta.KeyUserRole, "writer")
	return user
}
func newOwner() *meta.Meta {
	user := meta.NewMeta(ownerZid)
	user.Set(meta.KeyTitle, "Owner")
	user.Set(meta.KeyRole, meta.ValueRoleUser)
	user.Set(meta.KeyUserRole, "owner")
	return user
}
func newZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "Any Zettel")
	return m
}
func newPublicZettel() *meta.Meta {
	m := meta.NewMeta(visZid)
	m.Set(meta.KeyTitle, "Public Zettel")
	m.Set(meta.KeyVisibility, meta.ValueVisibilityPublic)
	return m
}
func newLoginZettel() *meta.Meta {
	m := meta.NewMeta(visZid)
	m.Set(meta.KeyTitle, "Login Zettel")
	m.Set(meta.KeyVisibility, meta.ValueVisibilityLogin)
	return m
}
func newOwnerZettel() *meta.Meta {
	m := meta.NewMeta(visZid)
	m.Set(meta.KeyTitle, "Owner Zettel")
	m.Set(meta.KeyVisibility, meta.ValueVisibilityOwner)
	return m
}
func newExpertZettel() *meta.Meta {
	m := meta.NewMeta(visZid)
	m.Set(meta.KeyTitle, "Expert Zettel")
	m.Set(meta.KeyVisibility, meta.ValueVisibilityExpert)
	return m
}
func newRoFalseZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "No r/o Zettel")
	m.Set(meta.KeyReadOnly, "false")
	return m
}
func newRoTrueZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "A r/o Zettel")
	m.Set(meta.KeyReadOnly, "true")
	return m
}
func newRoReaderZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "Reader r/o Zettel")
	m.Set(meta.KeyReadOnly, "reader")
	return m
}
func newRoWriterZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "Writer r/o Zettel")
	m.Set(meta.KeyReadOnly, "writer")
	return m
}
func newRoOwnerZettel() *meta.Meta {
	m := meta.NewMeta(zettelZid)
	m.Set(meta.KeyTitle, "Owner r/o Zettel")
	m.Set(meta.KeyReadOnly, "owner")
	return m
}
func newUserZettel() *meta.Meta {
	m := meta.NewMeta(userZid)
	m.Set(meta.KeyTitle, "Any User")
	m.Set(meta.KeyRole, meta.ValueRoleUser)
	return m
}
