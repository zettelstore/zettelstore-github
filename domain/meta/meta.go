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
	"bytes"
	"io"
	"regexp"
	"sort"
	"strings"

	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/input"
	"zettelstore.de/z/runes"
)

// Meta contains all meta-data of a zettel.
type Meta struct {
	Zid     id.Zid
	pairs   map[string]string
	frozen  bool
	YamlSep bool
}

// New creates a new chunk for storing meta-data
func New(zid id.Zid) *Meta {
	return &Meta{Zid: zid, pairs: make(map[string]string, 3)}
}

// Clone returns a new copy of the same meta data that is not frozen.
func (m *Meta) Clone() *Meta {
	return &Meta{
		Zid:     m.Zid,
		pairs:   m.Map(),
		frozen:  false,
		YamlSep: m.YamlSep,
	}
}

// Map returns a copy of the meta data as a string map.
func (m *Meta) Map() map[string]string {
	pairs := make(map[string]string, len(m.pairs))
	for k, v := range m.pairs {
		pairs[k] = v
	}
	return pairs
}

var reKey = regexp.MustCompile("^[0-9a-z][-0-9a-z]{0,254}$")

// KeyIsValid returns true, the the key is a valid string.
func KeyIsValid(key string) bool {
	return reKey.MatchString(key)
}

// Supported key types.
const (
	TypeBool       = 'b'
	TypeCredential = 'c'
	TypeEmpty      = 'e'
	TypeID         = 'i'
	TypeNumber     = 'n'
	TypeString     = 's'
	TypeTagSet     = 'T'
	TypeURL        = 'u'
	TypeUnknown    = '\000'
	TypeWord       = 'w'
	TypeWordSet    = 'W'
)

// Predefined keys.
const (
	KeyID                = "id"
	KeyTitle             = "title"
	KeyRole              = "role"
	KeyTags              = "tags"
	KeySyntax            = "syntax"
	KeyCopyright         = "copyright"
	KeyCredential        = "credential"
	KeyDefaultCopyright  = "default-copyright"
	KeyDefaultLang       = "default-lang"
	KeyDefaultLicense    = "default-license"
	KeyDefaultRole       = "default-role"
	KeyDefaultSyntax     = "default-syntax"
	KeyDefaultTitle      = "default-title"
	KeyDefaultVisibility = "default-visibility"
	KeyDuplicates        = "duplicates"
	KeyExpertMode        = "expert-mode"
	KeyFooterHTML        = "footer-html"
	KeyLang              = "lang"
	KeyLicense           = "license"
	KeyListPageSize      = "list-page-size"
	KeyNewRole           = "new-role"
	KeyMarkerExternal    = "marker-external"
	KeyReadOnly          = "read-only"
	KeySiteName          = "site-name"
	KeyStart             = "start"
	KeyURL               = "url"
	KeyUserID            = "user-id"
	KeyUserRole          = "user-role"
	KeyVisibility        = "visibility"
	KeyYAMLHeader        = "yaml-header"
	KeyZettelFileSyntax  = "zettel-file-syntax"
)

var keyTypeMap = map[string]byte{
	KeyID:                TypeID,
	KeyTitle:             TypeString,
	KeyRole:              TypeWord,
	KeyTags:              TypeTagSet,
	KeySyntax:            TypeWord,
	KeyCopyright:         TypeString,
	KeyCredential:        TypeCredential,
	KeyDefaultCopyright:  TypeString,
	KeyDefaultLicense:    TypeEmpty,
	KeyDefaultLang:       TypeWord,
	KeyDefaultRole:       TypeWord,
	KeyDefaultSyntax:     TypeWord,
	KeyDefaultTitle:      TypeString,
	KeyDefaultVisibility: TypeWord,
	KeyDuplicates:        TypeBool,
	KeyExpertMode:        TypeBool,
	KeyFooterHTML:        TypeString,
	KeyUserID:            TypeWord,
	KeyLang:              TypeWord,
	KeyLicense:           TypeEmpty,
	KeyListPageSize:      TypeNumber,
	KeyNewRole:           TypeWord,
	KeyMarkerExternal:    TypeEmpty,
	KeyReadOnly:          TypeWord,
	KeySiteName:          TypeString,
	KeyStart:             TypeID,
	KeyURL:               TypeURL,
	KeyUserRole:          TypeWord,
	KeyVisibility:        TypeWord,
	KeyYAMLHeader:        TypeBool,
	KeyZettelFileSyntax:  TypeWordSet,
}

// Important values for some keys.
const (
	ValueRoleUser         = "user"
	ValueRoleNewTemplate  = "new-template"
	ValueTrue             = "true"
	ValueFalse            = "false"
	ValueVisibilityExpert = "expert"
	ValueVisibilityOwner  = "owner"
	ValueVisibilityLogin  = "login"
	ValueVisibilityPublic = "public"
)

// Type returns a type hint for the given key. If no type hint is specified,
// TypeUnknown is returned.
func (m *Meta) Type(key string) byte {
	return KeyType(key)
}

// KeyType returns a type hint for the given key. If no type hint is specified,
// TypeUnknown is returned.
func KeyType(key string) byte {
	if t, ok := keyTypeMap[key]; ok {
		return t
	}
	return TypeUnknown
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

// Pair is one key-value-pair of a Zettel meta.
type Pair struct {
	Key   string
	Value string
}

var firstKeys = []string{KeyTitle, KeyRole, KeyTags, KeySyntax}
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
	if key != KeyID {
		m.pairs[key] = value
	}
}

// SetList stores the given string list value under the given key.
func (m *Meta) SetList(key string, values []string) {
	if m.frozen {
		panic("frozen.SetList")
	}
	if key != KeyID {
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
	if key == KeyID {
		return m.Zid.Format(), true
	}
	value, ok := m.pairs[key]
	return strings.TrimSpace(value), ok
}

// GetDefault retrieves the string value of the given key. If no value was
// stored, the given default value is returned.
func (m *Meta) GetDefault(key string, def string) string {
	if value, ok := m.Get(key); ok {
		return strings.TrimSpace(value)
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
func (m *Meta) Pairs() []Pair {
	return m.doPairs(true)
}

// PairsRest returns all key/values pairs stored, except the values with
// predefined keys. The pairs are ordered by key.
func (m *Meta) PairsRest() []Pair {
	return m.doPairs(false)
}

func (m *Meta) doPairs(first bool) []Pair {
	result := make([]Pair, 0, len(m.pairs))
	if first {
		for _, key := range firstKeys {
			if value, ok := m.pairs[key]; ok {
				result = append(result, Pair{key, strings.TrimSpace(value)})
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
		result = append(result, Pair{k, strings.TrimSpace(m.pairs[k])})
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
	if key != KeyID {
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

// NewFromInput parses the meta data of a zettel.
func NewFromInput(zid id.Zid, inp *input.Input) *Meta {
	if inp.Ch == '-' && inp.PeekN(0) == '-' && inp.PeekN(1) == '-' {
		skipToEOL(inp)
		inp.EatEOL()
	}
	meta := New(zid)
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
	case "", KeyID:
		// Empty key and 'id' key will be ignored
		return
	}

	switch KeyType(key) {
	case TypeString:
		if v != "" {
			addData(m, key, v)
		}
	case TypeTagSet:
		addSet(m, key, v, func(s string) bool { return s[0] == '#' })
	case TypeWord:
		m.Set(key, strings.ToLower(v))
	case TypeWordSet:
		addSet(m, key, strings.ToLower(v), func(s string) bool { return true })
	case TypeID:
		if _, err := id.Parse(val); err == nil {
			m.Set(key, val)
		}
	case TypeEmpty:
		fallthrough
	default:
		addData(m, key, v)
	}
}