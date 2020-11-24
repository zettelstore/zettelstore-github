//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package place provides a generic interface to zettel places.
package place

import (
	"strings"

	"zettelstore.de/z/domain"
)

// FilterFunc is a predicate to check if given meta must be selected.
type FilterFunc func(*domain.Meta) bool

func selectAll(m *domain.Meta) bool { return true }

type matchFunc func(value string) bool

func matchAlways(value string) bool { return true }
func matchNever(value string) bool  { return false }

type matchSpec struct {
	key   string
	match matchFunc
}

// CreateFilterFunc calculates a filter func based on the given filter.
func CreateFilterFunc(filter *Filter) FilterFunc {
	if filter == nil {
		return selectAll
	}
	specs := make([]matchSpec, 0, len(filter.Expr))
	var searchAll FilterFunc
	for key, values := range filter.Expr {
		if len(key) == 0 {
			// Special handling if searching all keys...
			searchAll = createSearchAllFunc(values, filter.Negate)
			continue
		}
		if domain.KeyIsValid(key) {
			match := createMatchFunc(key, values)
			if match != nil {
				specs = append(specs, matchSpec{key, match})
			}
		}
	}
	if len(specs) == 0 {
		if searchAll == nil {
			return selectAll
		}
		return searchAll
	}
	negate := filter.Negate
	searchMeta := func(m *domain.Meta) bool {
		for _, s := range specs {
			value, ok := m.Get(s.key)
			if !ok || !s.match(value) {
				return negate
			}
		}
		return !negate
	}
	if searchAll == nil {
		return searchMeta
	}
	return func(meta *domain.Meta) bool {
		return searchAll(meta) || searchMeta(meta)
	}
}

func createMatchFunc(key string, values []string) matchFunc {
	switch domain.KeyType(key) {
	case domain.MetaTypeBool:
		preValues := make([]bool, 0, len(values))
		for _, v := range values {
			preValues = append(preValues, domain.BoolValue(v))
		}
		return func(value string) bool {
			bValue := domain.BoolValue(value)
			for _, v := range preValues {
				if bValue != v {
					return false
				}
			}
			return true
		}
	case domain.MetaTypeCredential:
		return matchNever
	case domain.MetaTypeID:
		return func(value string) bool {
			for _, v := range values {
				if !strings.HasPrefix(value, v) {
					return false
				}
			}
			return true
		}
	case domain.MetaTypeTagSet:
		tagValues := preprocessSet(values)
		return func(value string) bool {
			tags := domain.ListFromValue(value)
			for _, neededTags := range tagValues {
				for _, neededTag := range neededTags {
					if !matchAllTag(tags, neededTag) {
						return false
					}
				}
			}
			return true
		}
	case domain.MetaTypeWord:
		values = sliceToLower(values)
		return func(value string) bool {
			value = strings.ToLower(value)
			for _, v := range values {
				if value != v {
					return false
				}
			}
			return true
		}
	case domain.MetaTypeWordSet:
		wordValues := preprocessSet(sliceToLower(values))
		return func(value string) bool {
			words := domain.ListFromValue(value)
			for _, neededWords := range wordValues {
				for _, neededWord := range neededWords {
					if !matchAllWord(words, neededWord) {
						return false
					}
				}
			}
			return true
		}
	}

	values = sliceToLower(values)
	return func(value string) bool {
		value = strings.ToLower(value)
		for _, v := range values {
			if !strings.Contains(value, v) {
				return false
			}
		}
		return true
	}
}

func createSearchAllFunc(values []string, negate bool) FilterFunc {
	matchFuncs := map[byte]matchFunc{}
	return func(meta *domain.Meta) bool {
		for _, p := range meta.Pairs() {
			keyType := domain.KeyType(p.Key)
			match, ok := matchFuncs[keyType]
			if !ok {
				match = createMatchFunc(p.Key, values)
				matchFuncs[keyType] = match
			}
			if match(p.Value) {
				return !negate
			}
		}
		match, ok := matchFuncs[domain.KeyType(domain.MetaKeyID)]
		if !ok {
			match = createMatchFunc(domain.MetaKeyID, values)
		}
		return match(meta.Zid.Format()) != negate
	}
}

func sliceToLower(sl []string) []string {
	result := make([]string, 0, len(sl))
	for _, s := range sl {
		result = append(result, strings.ToLower(s))
	}
	return result
}

func isEmptySlice(sl []string) bool {
	for _, s := range sl {
		if len(s) > 0 {
			return false
		}
	}
	return true
}

func preprocessSet(set []string) [][]string {
	result := make([][]string, 0, len(set))
	for _, elem := range set {
		splitElems := strings.Split(elem, ",")
		valueElems := make([]string, 0, len(splitElems))
		for _, se := range splitElems {
			e := strings.TrimSpace(se)
			if len(e) > 0 {
				valueElems = append(valueElems, e)
			}
		}
		if len(valueElems) > 0 {
			result = append(result, valueElems)
		}
	}
	return result
}

func matchAllTag(zettelTags []string, neededTag string) bool {
	for _, zt := range zettelTags {
		if zt == neededTag {
			return true
		}
	}
	return false
}

func matchAllWord(zettelWords []string, neededWord string) bool {
	for _, zw := range zettelWords {
		if zw == neededWord {
			return true
		}
	}
	return false
}
