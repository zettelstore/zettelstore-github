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

// Package domain provides domain specific types, constants, and functions.
package domain

import (
	"strings"
	"testing"

	"zettelstore.de/z/input"
)

const testID = ZettelID("98765432101234")

func newMeta(title string, tags []string, syntax string) *Meta {
	m := NewMeta(testID)
	if title != "" {
		m.Set(MetaKeyTitle, title)
	}
	if tags != nil {
		m.Set(MetaKeyTags, strings.Join(tags, " "))
	}
	if syntax != "" {
		m.Set(MetaKeySyntax, syntax)
	}
	return m
}

func TestKeyIsValid(t *testing.T) {
	validKeys := []string{"0", "a", "0-", "title", "title-----", strings.Repeat("r", 255)}
	for _, key := range validKeys {
		if !KeyIsValid(key) {
			t.Errorf("Key %q wrongly identified as invalid key", key)
		}
	}
	invalidKeys := []string{"", "-", "-a", "Title", "a_b", strings.Repeat("e", 256)}
	for _, key := range invalidKeys {
		if KeyIsValid(key) {
			t.Errorf("Key %q wrongly identified as valid key", key)
		}
	}
}

func assertWriteMeta(t *testing.T, meta *Meta, expected string) {
	t.Helper()
	sb := strings.Builder{}
	meta.Write(&sb)
	if got := sb.String(); got != expected {
		t.Errorf("\nExp: %q\ngot: %q", expected, got)
	}
}

func TestWriteMeta(t *testing.T) {
	assertWriteMeta(t, newMeta("", nil, ""), "")

	m := newMeta("TITLE", []string{"#t1", "#t2"}, "syntax")
	assertWriteMeta(t, m, "title: TITLE\ntags: #t1 #t2\nsyntax: syntax\n")

	m = newMeta("TITLE", nil, "")
	m.Set("user", "zettel")
	m.Set("auth", "basic")
	assertWriteMeta(t, m, "title: TITLE\nauth: basic\nuser: zettel\n")
}

func TestTitleHeader(t *testing.T) {
	m := NewMeta(testID)
	if got, ok := m.Get(MetaKeyTitle); ok || got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	addToMeta(m, MetaKeyTitle, " ")
	if got, ok := m.Get(MetaKeyTitle); ok || got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	const st = "A simple text"
	addToMeta(m, MetaKeyTitle, " "+st+"  ")
	if got, ok := m.Get(MetaKeyTitle); !ok || got != st {
		t.Errorf("Title is not %q, but %q", st, got)
	}
	addToMeta(m, MetaKeyTitle, "  "+st+"\t")
	const exp = st + " " + st
	if got, ok := m.Get(MetaKeyTitle); !ok || got != exp {
		t.Errorf("Title is not %q, but %q", exp, got)
	}

	m = NewMeta(testID)
	const at = "A Title"
	addToMeta(m, MetaKeyTitle, at)
	addToMeta(m, MetaKeyTitle, " ")
	if got, ok := m.Get(MetaKeyTitle); !ok || got != at {
		t.Errorf("Title is not %q, but %q", at, got)
	}
}

func checkSet(t *testing.T, exp []string, m *Meta, key string) {
	t.Helper()
	got, _ := m.GetList(key)
	for i, tag := range exp {
		if i < len(got) {
			if tag != got[i] {
				t.Errorf("Pos=%d, expected %q, got %q", i, exp[i], got[i])
			}
		} else {
			t.Errorf("Expected %q, but is missing", exp[i])
		}
	}
	if len(exp) < len(got) {
		t.Errorf("Extra tags: %q", got[len(exp):])
	}
}

func TestTagsHeader(t *testing.T) {
	m := NewMeta(testID)
	checkSet(t, []string{}, m, MetaKeyTags)

	addToMeta(m, MetaKeyTags, "")
	checkSet(t, []string{}, m, MetaKeyTags)

	addToMeta(m, MetaKeyTags, "  #t1 #t2  #t3 #t4  ")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4"}, m, MetaKeyTags)

	addToMeta(m, MetaKeyTags, "#t5")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m, MetaKeyTags)

	addToMeta(m, MetaKeyTags, "t6")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m, MetaKeyTags)
}

func TestSyntax(t *testing.T) {
	m := NewMeta(testID)
	if got, ok := m.Get(MetaKeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, MetaKeySyntax, " ")
	if got, _ := m.Get(MetaKeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, MetaKeySyntax, "MarkDown")
	const exp = "markdown"
	if got, ok := m.Get(MetaKeySyntax); !ok || got != exp {
		t.Errorf("Syntax is not %q, but %q", exp, got)
	}
	addToMeta(m, MetaKeySyntax, " ")
	if got, _ := m.Get(MetaKeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
}

func checkHeader(t *testing.T, exp map[string]string, gotP []MetaPair) {
	t.Helper()
	got := make(map[string]string, len(gotP))
	for _, p := range gotP {
		got[p.Key] = p.Value
		if _, ok := exp[p.Key]; !ok {
			t.Errorf("Key %q is not expected, but has value %q", p.Key, p.Value)
		}
	}
	for k, v := range exp {
		if gv, ok := got[k]; !ok || v != gv {
			if ok {
				t.Errorf("Key %q is not %q, but %q", k, v, got[k])
			} else {
				t.Errorf("Key %q missing, should have value %q", k, v)
			}
		}
	}
}

func TestDefaultHeader(t *testing.T) {
	m := NewMeta(testID)
	addToMeta(m, "h1", "d1")
	addToMeta(m, "H2", "D2")
	addToMeta(m, "H1", "D1.1")
	exp := map[string]string{"h1": "d1 D1.1", "h2": "D2"}
	checkHeader(t, exp, m.Pairs())
	addToMeta(m, "", "d0")
	checkHeader(t, exp, m.Pairs())
	addToMeta(m, "h3", "")
	exp["h3"] = ""
	checkHeader(t, exp, m.Pairs())
	addToMeta(m, "h3", "  ")
	checkHeader(t, exp, m.Pairs())
	addToMeta(m, "h4", " ")
	exp["h4"] = ""
	checkHeader(t, exp, m.Pairs())
}

func parseMetaStr(src string) *Meta {
	return NewMetaFromInput(testID, input.NewInput(src))
}

func TestEmpty(t *testing.T) {
	m := parseMetaStr("")
	if got, ok := m.Get(MetaKeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	if got, ok := m.GetList(MetaKeyTags); ok || len(got) > 0 {
		t.Errorf("Tags are not nil, but %v", got)
	}
}

func TestTitle(t *testing.T) {
	td := []struct{ s, e string }{
		{MetaKeyTitle + ": a title", "a title"},
		{MetaKeyTitle + ": a\n\t title", "a title"},
		{MetaKeyTitle + ": a\n\t title\r\n  x", "a title x"},
		{MetaKeyTitle + " AbC", "AbC"},
		{MetaKeyTitle + " AbC\n ded", "AbC ded"},
		{MetaKeyTitle + ": o\ntitle: p", "o p"},
		{MetaKeyTitle + ": O\n\ntitle: P", "O"},
		{MetaKeyTitle + ": b\r\ntitle: c", "b c"},
		{MetaKeyTitle + ": B\r\n\r\ntitle: C", "B"},
		{MetaKeyTitle + ": r\rtitle: q", "r q"},
		{MetaKeyTitle + ": R\r\rtitle: Q", "R"},
	}
	for i, tc := range td {
		m := parseMetaStr(tc.s)
		if got, ok := m.Get(MetaKeyTitle); !ok || got != tc.e {
			t.Log(m)
			t.Errorf("TC=%d: expected %q, got %q", i, tc.e, got)
		}
	}
}

func TestDelete(t *testing.T) {
	m := NewMeta(testID)
	m.Set("key", "val")
	if got, ok := m.Get("key"); !ok || got != "val" {
		t.Errorf("Value != %q, got: %v/%q", "val", ok, got)
	}
	m.Set("key", "")
	if got, ok := m.Get("key"); !ok || got != "" {
		t.Errorf("Value != %q, got: %v/%q", "", ok, got)
	}
	m.Delete("key")
	if got, ok := m.Get("key"); ok || got != "" {
		t.Errorf("Value != %q, got: %v/%q", "", ok, got)
	}
}

func TestNewMetaFromInput(t *testing.T) {
	testcases := []struct {
		input string
		exp   []MetaPair
	}{
		{"", []MetaPair{}},
		{" a:b", []MetaPair{{"a", "b"}}},
		{"%a:b", []MetaPair{}},
		{"a:b\r\n\r\nc:d", []MetaPair{{"a", "b"}}},
		{"a:b\r\n%c:d", []MetaPair{{"a", "b"}}},
		{"% a:b\r\n c:d", []MetaPair{{"c", "d"}}},
		{"---\r\na:b\r\n", []MetaPair{{"a", "b"}}},
		{"---\r\na:b\r\n--\r\nc:d", []MetaPair{{"a", "b"}, {"c", "d"}}},
		{"---\r\na:b\r\n---\r\nc:d", []MetaPair{{"a", "b"}}},
		{"---\r\na:b\r\n----\r\nc:d", []MetaPair{{"a", "b"}}},
	}
	for i, tc := range testcases {
		meta := parseMetaStr(tc.input)
		if got := meta.Pairs(); !equalPairs(tc.exp, got) {
			t.Errorf("TC=%d: expected=%v, got=%v", i, tc.exp, got)
		}
	}

	// Test, whether input position is correct.
	inp := input.NewInput("---\na:b\n---\nX")
	meta := NewMetaFromInput(testID, inp)
	exp := []MetaPair{{"a", "b"}}
	if got := meta.Pairs(); !equalPairs(exp, got) {
		t.Errorf("Expected=%v, got=%v", exp, got)
	}
	expCh := 'X'
	if gotCh := inp.Ch; gotCh != expCh {
		t.Errorf("Expected=%v, got=%v", expCh, gotCh)
	}
}

func equalPairs(one, two []MetaPair) bool {
	if len(one) != len(two) {
		return false
	}
	for i := 0; i < len(one); i++ {
		if one[i].Key != two[i].Key || one[i].Value != two[i].Value {
			return false
		}
	}
	return true
}
