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
	"regexp"
	"time"
)

// ZettelID ------------------------------------------------------------------

// ZettelID is the identification of a zettel. Typically, it is a time stamp in
// the form "YYYYMMDDHHmmSS". The store tries to set the last two digits to
// "00".
type ZettelID string

// RegexpID contains the regexp string that determines a valid zettel id.
const RegexpID = "[0-9]{14}"

var reID = regexp.MustCompile("^" + RegexpID + "$")

// IsValidID returns true, if string is a valid zettel ID.
func IsValidID(s string) bool {
	return reID.MatchString(s)
}

// IsValid returns true, the the zettel id is a valid string.
func (id ZettelID) IsValid() bool {
	return IsValidID(string(id))
}

// NewZettelID returns a new zettel ID based on the current time.
func NewZettelID(withSeconds bool) ZettelID {
	now := time.Now()
	if withSeconds {
		return ZettelID(now.Format("20060102150405"))
	}
	return ZettelID(now.Format("20060102150400"))
}

// Some important ZettelIDs
const (
	ConfigurationID  = ZettelID("00000000000000")
	BaseTemplateID   = ZettelID("00000000010100")
	LoginTemplateID  = ZettelID("00000000010200")
	ListTemplateID   = ZettelID("00000000010300")
	DetailTemplateID = ZettelID("00000000010401")
	InfoTemplateID   = ZettelID("00000000010402")
	FormTemplateID   = ZettelID("00000000010403")
	RenameTemplateID = ZettelID("00000000010404")
	DeleteTemplateID = ZettelID("00000000010405")
	RolesTemplateID  = ZettelID("00000000010500")
	TagsTemplateID   = ZettelID("00000000010600")
	BaseCSSID        = ZettelID("00000000020001")
	MaterialIconID   = ZettelID("00000000030001")
)

// Content -------------------------------------------------------------------

// Content is just the uninterpreted content of a zettel.
type Content string

// NewContent creates a new content from a string.
func NewContent(s string) Content { return Content(s) }

// AsString returns the content itself is a string.
func (zc Content) AsString() string { return string(zc) }

// AsBytes returns the content itself is a byte slice.
func (zc Content) AsBytes() []byte { return []byte(zc) }

// Zettel --------------------------------------------------------------------

// Zettel is the main data object of a zettelstore.
type Zettel struct {
	Meta    *Meta   // Some additional meta-data.
	Content Content // The content of the zettel itself.
}

// Equal compares two zettel for equality.
func (z Zettel) Equal(o Zettel) bool {
	return z.Meta.Equal(o.Meta) && z.Content == o.Content
}
