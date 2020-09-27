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

// Package place provides a generic interface to zettel places.
package place

import (
	"sort"

	"zettelstore.de/z/domain"
)

// ApplySorter applies the given sorter to the slide of meta data.
func ApplySorter(metaList []*domain.Meta, s *Sorter) []*domain.Meta {
	if len(metaList) == 0 {
		return metaList
	}

	if s == nil {
		sort.Slice(metaList, func(i, j int) bool { return metaList[i].Zid > metaList[j].Zid })
		return metaList
	}

	sorter := getSortFunc(s.Order, s.Descending, metaList)
	sort.Slice(metaList, sorter)
	if s.Offset > 0 {
		if s.Offset > len(metaList) {
			return nil
		}
		metaList = metaList[s.Offset:]
	}
	if s.Limit > 0 && s.Limit < len(metaList) {
		metaList = metaList[:s.Limit]
	}
	return metaList
}

type sortFunc func(i, j int) bool

func getSortFunc(key string, descending bool, ml []*domain.Meta) sortFunc {
	keyType := domain.KeyType(key)
	if key == domain.MetaKeyID || keyType == domain.MetaTypeCred {
		if descending {
			return func(i, j int) bool { return ml[i].Zid > ml[j].Zid }
		}
		return func(i, j int) bool { return ml[i].Zid < ml[j].Zid }
	} else if keyType == domain.MetaTypeBool {
		if descending {
			return func(i, j int) bool {
				left := ml[i].GetBool(key)
				if left == ml[j].GetBool(key) {
					return i > j
				}
				return left
			}
		}
		return func(i, j int) bool {
			right := ml[j].GetBool(key)
			if ml[i].GetBool(key) == right {
				return i < j
			}
			return right
		}
	}

	if descending {
		return func(i, j int) bool {
			iVal, iOk := ml[i].Get(key)
			jVal, jOk := ml[j].Get(key)
			return (iOk && (!jOk || iVal > jVal)) || !jOk
		}
	}
	return func(i, j int) bool {
		iVal, iOk := ml[i].Get(key)
		jVal, jOk := ml[j].Get(key)
		return (iOk && (!jOk || iVal < jVal)) || !jOk
	}
}
