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

import "strings"

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

// SetList stores the given string list value under the given key.
func (m *Meta) SetList(key string, values []string) {
	if m.frozen {
		panic("frozen.SetList")
	}
	if key != KeyID {
		m.pairs[key] = strings.Join(values, " ")
	}
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
