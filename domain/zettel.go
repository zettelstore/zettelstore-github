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
	"strconv"
	"time"
)

// ZettelID ------------------------------------------------------------------

// ZettelID is the internal identification of a zettel. Typically, it is a
// time stamp of the form "YYYYMMDDHHmmSS" converted to an unsigned integer.
// A zettelstore implementation should try to set the last two digits to zero,
// e.g. the seconds should be zero,
type ZettelID uint64

// InvalidZettelID is a ZettelID that will never be valid
const InvalidZettelID = ZettelID(0)

const maxZettelID = 99999999999999

// ParseZettelID interprets a string as a zettel identification and
// returns its integer value.
func ParseZettelID(s string) (ZettelID, error) {
	if len(s) != 14 {
		return InvalidZettelID, strconv.ErrSyntax
	}
	res, err := strconv.ParseUint(s, 10, 47)
	if err != nil {
		return InvalidZettelID, err
	}
	if res == 0 {
		return InvalidZettelID, strconv.ErrRange
	}
	return ZettelID(res), nil
}

const digits = "0123456789"

// Format converts the zettel identification to a string of 14 digits.
// Only defined for valid ids.
func (zid ZettelID) Format() string {
	return string(zid.FormatBytes())
}

// FormatBytes converts the zettel identification to a byte slice of 14 digits.
// Only defined for valid ids.
func (zid ZettelID) FormatBytes() []byte {
	result := make([]byte, 14)
	for i := 13; i >= 0; i-- {
		result[i] = digits[zid%10]
		zid /= 10
	}
	return result
}

// IsValid determines if zettel id is a valid one, e.g. consists of max. 14 digits.
func (zid ZettelID) IsValid() bool { return 0 < zid && zid <= maxZettelID }

// NewZettelID returns a new zettel id based on the current time.
func NewZettelID(withSeconds bool) ZettelID {
	now := time.Now()
	var s string
	if withSeconds {
		s = now.Format("20060102150405")
	} else {
		s = now.Format("20060102150400")
	}
	res, err := ParseZettelID(s)
	if err != nil {
		panic(err)
	}
	return res
}

// Some important ZettelIDs
const (
	ConfigurationID  = ZettelID(1)
	BaseTemplateID   = ZettelID(10100)
	LoginTemplateID  = ZettelID(10200)
	ListTemplateID   = ZettelID(10300)
	DetailTemplateID = ZettelID(10401)
	InfoTemplateID   = ZettelID(10402)
	FormTemplateID   = ZettelID(10403)
	RenameTemplateID = ZettelID(10404)
	DeleteTemplateID = ZettelID(10405)
	RolesTemplateID  = ZettelID(10500)
	TagsTemplateID   = ZettelID(10600)
	BaseCSSID        = ZettelID(20001)
	MaterialIconID   = ZettelID(30001)
	TemplateZettelID = ZettelID(40001)
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
