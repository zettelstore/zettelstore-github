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

// Package strfun provides some string functions.
package strfun

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	useUnicode = []*unicode.RangeTable{
		unicode.Letter,
		unicode.Number,
	}
	ignoreUnicode = []*unicode.RangeTable{
		unicode.Mark,
		unicode.Sk,
		unicode.Lm,
	}
)

// Slugify returns a string that can be used as part of an URL
func Slugify(s string) string {
	s = strings.TrimSpace(s)
	result := make([]rune, 0, len(s))
	addDash := false
	for _, r := range norm.NFKD.String(s) {
		if unicode.IsOneOf(useUnicode, r) {
			result = append(result, unicode.ToLower(r))
			addDash = true
		} else if !unicode.IsOneOf(ignoreUnicode, r) && addDash {
			result = append(result, '-')
			addDash = false
		}
	}
	if i := len(result) - 1; i >= 0 && result[i] == '-' {
		result = result[:i]
	}
	return string(result)
}
