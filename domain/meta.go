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
	"bytes"
	"io"
	"regexp"
	"sort"
	"strings"

	"zettelstore.de/z/input"
	"zettelstore.de/z/runes"
)

// Meta contains all meta-data of a zettel.
type Meta struct {
	Zid     ZettelID
	pairs   map[string]string
	frozen  bool
	YamlSep bool
}

// NewMeta creates a new chunk for storing meta-data
func NewMeta(zid ZettelID) *Meta {
	return &Meta{Zid: zid, pairs: make(map[string]string, 3)}
}

// Clone returns a new copy of the same meta data that is not frozen.
func (m *Meta) Clone() *Meta {
	pairs := make(map[string]string, len(m.pairs))
	for k, v := range m.pairs {
		pairs[k] = v
	}
	return &Meta{
		Zid:     m.Zid,
		pairs:   pairs,
		frozen:  false,
		YamlSep: m.YamlSep,
	}
}

var reKey = regexp.MustCompile("^[0-9a-z][-0-9a-z]{0,254}$")

// KeyIsValid returns true, the the key is a valid string.
func KeyIsValid(key string) bool {
	return reKey.MatchString(key)
}

// Predefined keys.
const (
	MetaKeyID               = "id"
	MetaKeyTitle            = "title"
	MetaKeyTags             = "tags"
	MetaKeySyntax           = "syntax"
	MetaKeyRole             = "role"
	MetaKeyCopyright        = "copyright"
	MetaKeyCred             = "cred"
	MetaKeyDefaultCopyright = "default-copyright"
	MetaKeyDefaultLang      = "default-lang"
	MetaKeyDefaultLicense   = "default-license"
	MetaKeyDefaultRole      = "default-role"
	MetaKeyDefaultSyntax    = "default-syntax"
	MetaKeyDefaultTitle     = "default-title"
	MetaKeyIconMaterial     = "icon-material"
	MetaKeyIdent            = "ident"
	MetaKeyLang             = "lang"
	MetaKeyLicense          = "license"
	MetaKeyOwner            = "owner"
	MetaKeySiteName         = "site-name"
	MetaKeyStart            = "start"
	MetaKeyURL              = "url"
	MetaKeyYAMLHeader       = "yaml-header"
	MetaKeyZettelFileSyntax = "zettel-file-syntax"
)

// Supported key types.
const (
	MetaTypeBool    = 'b'
	MetaTypeCred    = 'c'
	MetaTypeEmpty   = 'e'
	MetaTypeID      = 'i'
	MetaTypeString  = 's'
	MetaTypeTagSet  = 'T'
	MetaTypeURL     = 'u'
	MetaTypeUnknown = '\000'
	MetaTypeWord    = 'w'
	MetaTypeWordSet = 'W'
)

var keyTypeMap = map[string]byte{
	MetaKeyID:               MetaTypeID,
	MetaKeyTitle:            MetaTypeString,
	MetaKeyTags:             MetaTypeTagSet,
	MetaKeySyntax:           MetaTypeWord,
	MetaKeyRole:             MetaTypeWord,
	MetaKeyCopyright:        MetaTypeString,
	MetaKeyCred:             MetaTypeCred,
	MetaKeyDefaultCopyright: MetaTypeString,
	MetaKeyDefaultLicense:   MetaTypeEmpty,
	MetaKeyDefaultLang:      MetaTypeWord,
	MetaKeyDefaultRole:      MetaTypeWord,
	MetaKeyDefaultSyntax:    MetaTypeWord,
	MetaKeyDefaultTitle:     MetaTypeString,
	MetaKeyIdent:            MetaTypeWord,
	MetaKeyLang:             MetaTypeWord,
	MetaKeyLicense:          MetaTypeEmpty,
	MetaKeyOwner:            MetaTypeID,
	MetaKeySiteName:         MetaTypeString,
	MetaKeyStart:            MetaTypeID,
	MetaKeyURL:              MetaTypeURL,
	MetaKeyYAMLHeader:       MetaTypeBool,
	MetaKeyZettelFileSyntax: MetaTypeWordSet,
}

// Type returns a type hint for the given key. If no type hint is specified,
// MetaTypeUnknown is returned.
func (m *Meta) Type(key string) byte {
	return KeyType(key)
}

// KeyType returns a type hint for the given key. If no type hint is specified,
// MetaTypeUnknown is returned.
func KeyType(key string) byte {
	if t, ok := keyTypeMap[key]; ok {
		return t
	}
	return MetaTypeUnknown
}

// BoolValue returns the value interpreted as a bool.
func BoolValue(value string) bool {
	if len(value) > 0 {
		switch value[0] {
		case '0', 'f', 'F', 'n', 'N':
			return false
		}
	}
	return true
}

// MetaPair is one key-value-pair of a Zettel meta.
type MetaPair struct {
	Key   string
	Value string
}

var firstKeys = []string{MetaKeyTitle, MetaKeyTags, MetaKeySyntax, MetaKeyRole}
var firstKeySet map[string]bool

func init() {
	firstKeySet = make(map[string]bool, len(firstKeys))
	for _, k := range firstKeys {
		firstKeySet[k] = true
	}
}

// Set stores the given string value under the given key.
func (m *Meta) Set(key, value string) {
	if m.frozen {
		panic("frozen.Set")
	}
	if key != MetaKeyID {
		m.pairs[key] = value
	}
}

// SetList stores the given string list value under the given key.
func (m *Meta) SetList(key string, values []string) {
	if m.frozen {
		panic("frozen.SetList")
	}
	if key != MetaKeyID {
		m.pairs[key] = strings.Join(values, " ")
	}
}

// IsFrozen returns whether meta can be changed or not.
func (m *Meta) IsFrozen() bool { return m.frozen }

// Freeze defines frozen meta data, i.e. changing them will result in a panic.
func (m *Meta) Freeze() {
	m.frozen = true
}

// Get retrieves the string value of a given key. The bool value signals,
// whether there was a value stored or not.
func (m *Meta) Get(key string) (string, bool) {
	if key == MetaKeyID {
		return m.Zid.Format(), true
	}
	value, ok := m.pairs[key]
	return value, ok
}

// GetDefault retrieves the string value of the given key. If no value was
// stored, the given default value is returned.
func (m *Meta) GetDefault(key string, def string) string {
	if value, ok := m.Get(key); ok {
		return value
	}
	return def
}

// GetBool returns the boolean value of the given key.
func (m *Meta) GetBool(key string) bool {
	if value, ok := m.Get(key); ok {
		return BoolValue(value)
	}
	return false
}

// ListFromValue transforms a string value into a list value.
func ListFromValue(value string) []string {
	return strings.Fields(value)
}

// GetList retrieves the string list value of a given key. The bool value
// signals, whether there was a value stored or not.
func (m *Meta) GetList(key string) ([]string, bool) {
	value, ok := m.Get(key)
	if !ok {
		return nil, false
	}
	return ListFromValue(value), true
}

// GetListOrNil retrieves the string list value of a given key. If there was
// nothing stores, a nil list is returned.
func (m *Meta) GetListOrNil(key string) []string {
	if value, ok := m.GetList(key); ok {
		return value
	}
	return nil
}

// Pairs returns all key/values pairs stored, in a specific order. First come
// the pairs with predefined keys: MetaTitleKey, MetaTagsKey, MetaSyntaxKey,
// MetaContextKey. Then all other pairs are append to the list, ordered by key.
func (m *Meta) Pairs() []MetaPair {
	return m.doPairs(true)
}

// PairsRest returns all key/values pairs stored, except the values with
// predefined keys. The pairs are ordered by key.
func (m *Meta) PairsRest() []MetaPair {
	return m.doPairs(false)
}

func (m *Meta) doPairs(first bool) []MetaPair {
	result := make([]MetaPair, 0, len(m.pairs))
	if first {
		for _, key := range firstKeys {
			if value, ok := m.pairs[key]; ok {
				result = append(result, MetaPair{Key: key, Value: value})
			}
		}
	}

	keys := make([]string, 0, len(m.pairs)-len(result))
	for k := range m.pairs {
		if !firstKeySet[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, MetaPair{k, m.pairs[k]})
	}
	return result
}

// Write writes a zettel meta to a writer.
func (m *Meta) Write(w io.Writer) (int, error) {
	var buf bytes.Buffer
	for _, p := range m.Pairs() {
		buf.WriteString(p.Key)
		buf.WriteString(": ")
		buf.WriteString(p.Value)
		buf.WriteByte('\n')
	}
	return w.Write(buf.Bytes())
}

var (
	newline = []byte{'\n'}
	yamlSep = []byte{'-', '-', '-', '\n'}
)

// WriteAsHeader writes the zettel meta to the writer, plus the separators
func (m *Meta) WriteAsHeader(w io.Writer) (int, error) {
	var lb, lc, la int
	var err error

	if m.YamlSep {
		lb, err = w.Write(yamlSep)
		if err != nil {
			return lb, err
		}
	}
	lc, err = m.Write(w)
	if err != nil {
		return lb + lc, err
	}
	if m.YamlSep {
		la, err = w.Write(yamlSep)
	} else {
		la, err = w.Write(newline)
	}
	return lb + lc + la, err
}

// Delete removes a key from the data.
func (m *Meta) Delete(key string) {
	if m.frozen {
		panic("frozen.Delete")
	}
	if key != MetaKeyID {
		delete(m.pairs, key)
	}
}

// Equal compares to metas for equality.
func (m *Meta) Equal(o *Meta) bool {
	if m == nil && o == nil {
		return true
	}
	if m == nil || o == nil || m.Zid != o.Zid || len(m.pairs) != len(o.pairs) {
		return false
	}
	for k, v := range m.pairs {
		if vo, ok := o.pairs[k]; !ok || v != vo {
			return false
		}
	}
	return true
}

// NewMetaFromInput parses the meta data of a zettel.
func NewMetaFromInput(zid ZettelID, inp *input.Input) *Meta {
	if inp.Ch == '-' && inp.PeekN(0) == '-' && inp.PeekN(1) == '-' {
		skipToEOL(inp)
		inp.EatEOL()
	}
	meta := NewMeta(zid)
	for {
		skipSpace(inp)
		switch inp.Ch {
		case '\r':
			if inp.Peek() == '\n' {
				inp.Next()
			}
			fallthrough
		case '\n':
			inp.Next()
			return meta
		case input.EOS:
			return meta
		case '%':
			skipToEOL(inp)
			inp.EatEOL()
			continue
		}
		parseHeader(meta, inp)
		if inp.Ch == '-' && inp.PeekN(0) == '-' && inp.PeekN(1) == '-' {
			skipToEOL(inp)
			inp.EatEOL()
			meta.YamlSep = true
			return meta
		}
	}
}

func parseHeader(m *Meta, inp *input.Input) {
	pos := inp.Pos
	for isHeader(inp.Ch) {
		inp.Next()
	}
	key := inp.Src[pos:inp.Pos]
	skipSpace(inp)
	if inp.Ch == ':' {
		inp.Next()
	}
	var val string
	for {
		skipSpace(inp)
		pos = inp.Pos
		skipToEOL(inp)
		val += inp.Src[pos:inp.Pos]
		inp.EatEOL()
		if !runes.IsSpace(inp.Ch) {
			break
		}
		val += " "
	}
	addToMeta(m, key, val)
}

func skipSpace(inp *input.Input) {
	for runes.IsSpace(inp.Ch) {
		inp.Next()
	}
}

func skipToEOL(inp *input.Input) {
	for {
		switch inp.Ch {
		case '\n', '\r', input.EOS:
			return
		}
		inp.Next()
	}
}

// Return true iff rune is valid for header key.
func isHeader(ch rune) bool {
	return ('a' <= ch && ch <= 'z') ||
		('0' <= ch && ch <= '9') ||
		ch == '-' ||
		('A' <= ch && ch <= 'Z')
}

type predValidElem func(string) bool

func addToSet(set map[string]bool, elems []string, useElem predValidElem) {
	for _, s := range elems {
		if len(s) > 0 && useElem(s) {
			set[s] = true
		}
	}
}

func addSet(m *Meta, key, val string, useElem predValidElem) {
	newElems := strings.Fields(val)
	oldElems, ok := m.GetList(key)
	if !ok {
		oldElems = nil
	}

	set := make(map[string]bool, len(newElems)+len(oldElems))
	addToSet(set, newElems, useElem)
	if len(set) == 0 {
		// Nothing to add. Maybe because of filtered elements.
		return
	}
	addToSet(set, oldElems, useElem)

	resultList := make([]string, 0, len(set))
	for tag := range set {
		resultList = append(resultList, tag)
	}
	sort.Strings(resultList)
	m.SetList(key, resultList)
}

func addData(m *Meta, k, v string) {
	if o, ok := m.Get(k); !ok || o == "" {
		m.Set(k, v)
	} else if v != "" {
		m.Set(k, o+" "+v)
	}
}

func addToMeta(m *Meta, key, val string) {
	v := strings.TrimFunc(val, runes.IsSpace)
	key = strings.ToLower(key)
	if !KeyIsValid(key) {
		return
	}
	switch key {
	case "", MetaKeyID:
		// Empty key and 'id' key will be ignored
		return
	}

	switch KeyType(key) {
	case MetaTypeString:
		if v != "" {
			addData(m, key, v)
		}
	case MetaTypeTagSet:
		addSet(m, key, v, func(s string) bool { return s[0] == '#' })
	case MetaTypeWord:
		m.Set(key, strings.ToLower(v))
	case MetaTypeWordSet:
		addSet(m, key, strings.ToLower(v), func(s string) bool { return true })
	case MetaTypeID:
		if _, err := ParseZettelID(val); err == nil {
			m.Set(key, val)
		}
	//case MetaContextKey:
	//	addSet(m, key, v, func(s string) bool { return IsValidID(s) })
	case MetaTypeEmpty:
		fallthrough
	default:
		addData(m, key, v)
	}
}
