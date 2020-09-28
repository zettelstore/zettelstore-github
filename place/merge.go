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

import "zettelstore.de/z/domain"

// MergeSorted returns a merged sequence of meta data, sorted by a given Sorter.
// The lists first and second must be sorted descending by Zid.
func MergeSorted(first, second []*domain.Meta, s *Sorter) []*domain.Meta {
	lenFirst := len(first)
	lenSecond := len(second)
	result := make([]*domain.Meta, 0, lenFirst+lenSecond)
	iFirst := 0
	iSecond := 0
	for iFirst < lenFirst && iSecond < lenSecond {
		zidFirst := first[iFirst].Zid
		zidSecond := second[iSecond].Zid
		if zidFirst > zidSecond {
			result = append(result, first[iFirst])
			iFirst++
		} else if zidFirst < zidSecond {
			result = append(result, second[iSecond])
			iSecond++
		} else { // zidFirst == zidSecond
			result = append(result, first[iFirst])
			iFirst++
			iSecond++
		}
	}
	if iFirst < lenFirst {
		result = append(result, first[iFirst:]...)
	} else {
		result = append(result, second[iSecond:]...)
	}
	if s == nil {
		return result
	}
	return ApplySorter(result, s)
}
