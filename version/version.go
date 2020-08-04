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

// Package version provides version information.
package version

// Version describes all elements of a software version.
type Version struct {
	Release string // Official software release version
	Build   string // Internal representation of build process
	// More to come
}

var v Version

// Setup initializes the version data.
func Setup(release string, build string) {
	if len(release) > 0 {
		v.Release = release
	} else {
		v.Release = "unknown"
	}
	if len(build) > 0 {
		v.Build = build
	} else {
		v.Build = "unknown"
	}
}

// Get returns the current software version data.
func Get() Version {
	return v
}
