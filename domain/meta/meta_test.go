//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package meta provides the domain specific type 'meta'.
package meta

import (
	"strings"
	"testing"

	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/input"
)

const testID = id.Zid(98765432101234)

func newMeta(title string, tags []string, syntax string) *Meta {
	m := New(testID)
	if title != "" {
		m.Set(KeyTitle, title)
	}
	if tags != nil {
		m.Set(KeyTags, strings.Join(tags, " "))
	}
	if syntax != "" {
		m.Set(KeySyntax, syntax)
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
	m := New(testID)
	if got, ok := m.Get(KeyTitle); ok || got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	addToMeta(m, KeyTitle, " ")
	if got, ok := m.Get(KeyTitle); ok || got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	const st = "A simple text"
	addToMeta(m, KeyTitle, " "+st+"  ")
	if got, ok := m.Get(KeyTitle); !ok || got != st {
		t.Errorf("Title is not %q, but %q", st, got)
	}
	addToMeta(m, KeyTitle, "  "+st+"\t")
	const exp = st + " " + st
	if got, ok := m.Get(KeyTitle); !ok || got != exp {
		t.Errorf("Title is not %q, but %q", exp, got)
	}

	m = New(testID)
	const at = "A Title"
	addToMeta(m, KeyTitle, at)
	addToMeta(m, KeyTitle, " ")
	if got, ok := m.Get(KeyTitle); !ok || got != at {
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
	m := New(testID)
	checkSet(t, []string{}, m, KeyTags)

	addToMeta(m, KeyTags, "")
	checkSet(t, []string{}, m, KeyTags)

	addToMeta(m, KeyTags, "  #t1 #t2  #t3 #t4  ")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4"}, m, KeyTags)

	addToMeta(m, KeyTags, "#t5")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m, KeyTags)

	addToMeta(m, KeyTags, "t6")
	checkSet(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m, KeyTags)
}

func TestSyntax(t *testing.T) {
	m := New(testID)
	if got, ok := m.Get(KeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, KeySyntax, " ")
	if got, _ := m.Get(KeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, KeySyntax, "MarkDown")
	const exp = "markdown"
	if got, ok := m.Get(KeySyntax); !ok || got != exp {
		t.Errorf("Syntax is not %q, but %q", exp, got)
	}
	addToMeta(m, KeySyntax, " ")
	if got, _ := m.Get(KeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
}

func checkHeader(t *testing.T, exp map[string]string, gotP []Pair) {
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
	m := New(testID)
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
	return NewFromInput(testID, input.NewInput(src))
}

func TestEmpty(t *testing.T) {
	m := parseMetaStr("")
	if got, ok := m.Get(KeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	if got, ok := m.GetList(KeyTags); ok || len(got) > 0 {
		t.Errorf("Tags are not nil, but %v", got)
	}
}

func TestTitle(t *testing.T) {
	td := []struct{ s, e string }{
		{KeyTitle + ": a title", "a title"},
		{KeyTitle + ": a\n\t title", "a title"},
		{KeyTitle + ": a\n\t title\r\n  x", "a title x"},
		{KeyTitle + " AbC", "AbC"},
		{KeyTitle + " AbC\n ded", "AbC ded"},
		{KeyTitle + ": o\ntitle: p", "o p"},
		{KeyTitle + ": O\n\ntitle: P", "O"},
		{KeyTitle + ": b\r\ntitle: c", "b c"},
		{KeyTitle + ": B\r\n\r\ntitle: C", "B"},
		{KeyTitle + ": r\rtitle: q", "r q"},
		{KeyTitle + ": R\r\rtitle: Q", "R"},
	}
	for i, tc := range td {
		m := parseMetaStr(tc.s)
		if got, ok := m.Get(KeyTitle); !ok || got != tc.e {
			t.Log(m)
			t.Errorf("TC=%d: expected %q, got %q", i, tc.e, got)
		}
	}
}

func TestDelete(t *testing.T) {
	m := New(testID)
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

func TestNewFromInput(t *testing.T) {
	testcases := []struct {
		input string
		exp   []Pair
	}{
		{"", []Pair{}},
		{" a:b", []Pair{{"a", "b"}}},
		{"%a:b", []Pair{}},
		{"a:b\r\n\r\nc:d", []Pair{{"a", "b"}}},
		{"a:b\r\n%c:d", []Pair{{"a", "b"}}},
		{"% a:b\r\n c:d", []Pair{{"c", "d"}}},
		{"---\r\na:b\r\n", []Pair{{"a", "b"}}},
		{"---\r\na:b\r\n--\r\nc:d", []Pair{{"a", "b"}, {"c", "d"}}},
		{"---\r\na:b\r\n---\r\nc:d", []Pair{{"a", "b"}}},
		{"---\r\na:b\r\n----\r\nc:d", []Pair{{"a", "b"}}},
	}
	for i, tc := range testcases {
		meta := parseMetaStr(tc.input)
		if got := meta.Pairs(); !equalPairs(tc.exp, got) {
			t.Errorf("TC=%d: expected=%v, got=%v", i, tc.exp, got)
		}
	}

	// Test, whether input position is correct.
	inp := input.NewInput("---\na:b\n---\nX")
	meta := NewFromInput(testID, inp)
	exp := []Pair{{"a", "b"}}
	if got := meta.Pairs(); !equalPairs(exp, got) {
		t.Errorf("Expected=%v, got=%v", exp, got)
	}
	expCh := 'X'
	if gotCh := inp.Ch; gotCh != expCh {
		t.Errorf("Expected=%v, got=%v", expCh, gotCh)
	}
}

func equalPairs(one, two []Pair) bool {
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