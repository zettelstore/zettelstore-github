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
	"unicode/utf8"
)

// Content is just the uninterpreted content of a zettel.
type Content string

// NewContent creates a new content from a string.
func NewContent(s string) Content { return Content(s) }

// AsString returns the content itself is a string.
func (zc Content) AsString() string { return string(zc) }

// AsBytes returns the content itself is a byte slice.
func (zc Content) AsBytes() []byte { return []byte(zc) }

// IsBinary returns true if the content contains non-unicode values or is,
// interpreted a text, with a high probability binary content.
func (zc Content) IsBinary() bool {
	s := string(zc)
	if !utf8.ValidString(s) {
		return true
	}
	l := len(s)
	for i := 0; i < l; i++ {
		if s[i] == 0 {
			return true
		}
	}
	return false
}
