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
	KeyModified          = "modified"
	KeyPrecursor         = "precursor"
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
	KeyModified:          TypeDatetime,
	KeyPrecursor:         TypeID,
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
		m.pairs[key] = value
	}
}

// Get retrieves the string value of a given key. The bool value signals,
// whether there was a value stored or not.
func (m *Meta) Get(key string) (string, bool) {
	if key == KeyID {
		return m.Zid.String(), true
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
