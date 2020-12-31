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
	"regexp"
	"sort"
	"strings"

	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/runes"
)

// DescriptionKey formally describes each supported metadata key.
type DescriptionKey struct {
	Name       string
	Type       *DescriptionType
	IsComputed bool
}

var registeredKeys = make(map[string]*DescriptionKey)

func registerKey(name string, t *DescriptionType, isComputed bool) string {
	if _, ok := registeredKeys[name]; ok {
		panic("Key '" + name + "' already defined")
	}
	registeredKeys[name] = &DescriptionKey{name, t, isComputed}
	return name
}

// GetSortedKeyDescriptions delivers all metadata key descriptions as a slice, sorted by name.
func GetSortedKeyDescriptions() []*DescriptionKey {
	names := make([]string, 0, len(registeredKeys))
	for n := range registeredKeys {
		names = append(names, n)
	}
	sort.Strings(names)
	result := make([]*DescriptionKey, 0, len(names))
	for _, n := range names {
		result = append(result, registeredKeys[n])
	}
	return result
}

// Supported keys.
var (
	KeyID                = registerKey("id", TypeID, true)
	KeyTitle             = registerKey("title", TypeString, false)
	KeyRole              = registerKey("role", TypeWord, false)
	KeyTags              = registerKey("tags", TypeTagSet, false)
	KeySyntax            = registerKey("syntax", TypeWord, false)
	KeyCopyright         = registerKey("copyright", TypeString, false)
	KeyCredential        = registerKey("credential", TypeCredential, false)
	KeyDefaultCopyright  = registerKey("default-copyright", TypeString, false)
	KeyDefaultLang       = registerKey("default-lang", TypeWord, false)
	KeyDefaultLicense    = registerKey("default-license", TypeEmpty, false)
	KeyDefaultRole       = registerKey("default-role", TypeWord, false)
	KeyDefaultSyntax     = registerKey("default-syntax", TypeWord, false)
	KeyDefaultTitle      = registerKey("default-title", TypeString, false)
	KeyDefaultVisibility = registerKey("default-visibility", TypeWord, false)
	KeyDuplicates        = registerKey("duplicates", TypeBool, false)
	KeyExpertMode        = registerKey("expert-mode", TypeBool, false)
	KeyFooterHTML        = registerKey("footer-html", TypeString, false)
	KeyLang              = registerKey("lang", TypeWord, false)
	KeyLicense           = registerKey("license", TypeEmpty, false)
	KeyListPageSize      = registerKey("list-page-size", TypeNumber, false)
	KeyNewRole           = registerKey("new-role", TypeWord, false)
	KeyMarkerExternal    = registerKey("marker-external", TypeEmpty, false)
	KeyModified          = registerKey("modified", TypeTimestamp, true)
	KeyPrecursor         = registerKey("precursor", TypeID, false)
	KeyReadOnly          = registerKey("read-only", TypeWord, false)
	KeySiteName          = registerKey("site-name", TypeString, false)
	KeyStart             = registerKey("start", TypeID, false)
	KeyURL               = registerKey("url", TypeURL, false)
	KeyUserID            = registerKey("user-id", TypeWord, false)
	KeyUserRole          = registerKey("user-role", TypeWord, false)
	KeyVisibility        = registerKey("visibility", TypeWord, false)
	KeyYAMLHeader        = registerKey("yaml-header", TypeBool, false)
	KeyZettelFileSyntax  = registerKey("zettel-file-syntax", TypeWordSet, false)
)

// Important values for some keys.
const (
	ValueRoleConfiguration = "configuration"
	ValueRoleUser          = "user"
	ValueRoleNewTemplate   = "new-template"
	ValueRoleZettel        = "zettel"
	ValueSyntaxMeta        = "meta"
	ValueSyntaxZmk         = "zmk"
	ValueTrue              = "true"
	ValueFalse             = "false"
	ValueUserRoleReader    = "reader"
	ValueUserRoleWriter    = "writer"
	ValueUserRoleOwner     = "owner"
	ValueVisibilityExpert  = "expert"
	ValueVisibilityOwner   = "owner"
	ValueVisibilityLogin   = "login"
	ValueVisibilityPublic  = "public"
	ValueVisibilitySimple  = "simple-expert"
)

// Meta contains all meta-data of a zettel.
type Meta struct {
	Zid     id.Zid
	pairs   map[string]string
	YamlSep bool
}

// New creates a new chunk for storing meta-data
func New(zid id.Zid) *Meta {
	return &Meta{Zid: zid, pairs: make(map[string]string, 5)}
}

// Clone returns a new copy of the metadata.
func (m *Meta) Clone() *Meta {
	return &Meta{
		Zid:     m.Zid,
		pairs:   m.Map(),
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
	if key != KeyID {
		m.pairs[key] = trimValue(value)
	}
}

func trimValue(value string) string {
	return strings.TrimFunc(value, runes.IsSpace)
}

// Get retrieves the string value of a given key. The bool value signals,
// whether there was a value stored or not.
func (m *Meta) Get(key string) (string, bool) {
	if key == KeyID {
		return m.Zid.String(), true
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
				result = append(result, Pair{key, value})
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
		result = append(result, Pair{k, m.pairs[k]})
	}
	return result
}

// Delete removes a key from the data.
func (m *Meta) Delete(key string) {
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
