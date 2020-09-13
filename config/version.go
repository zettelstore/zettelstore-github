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

// Package config provides functions to retrieve configuration data.
package config

import (
	"os"
	"runtime"
)

// Version describes all elements of a software version.
type Version struct {
	Prog      string // Name of the software
	Build     string // Representation of build process
	Hostname  string // Host name a reported by the kernel
	GoVersion string // Version of go
	Os        string // GOOS
	Arch      string // GOARCH
	// More to come
}

var version Version

// SetupVersion initializes the version data.
func SetupVersion(progName, buildVersion string) {
	version.Prog = progName
	if buildVersion == "" {
		version.Build = "unknown"
	} else {
		version.Build = buildVersion
	}
	if hn, err := os.Hostname(); err == nil {
		version.Hostname = hn
	} else {
		version.Hostname = "*unknown host*"
	}
	version.GoVersion = runtime.Version()
	version.Os = runtime.GOOS
	version.Arch = runtime.GOARCH
}

// GetVersion returns the current software version data.
func GetVersion() Version { return version }
