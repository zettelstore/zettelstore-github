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

// Package cmd provides command generic functions.
package cmd

// Mention all needed encoders, parsers and stores to have them registered.
import (
	_ "zettelstore.de/z/encoder/htmlenc"   // Allow to use HTML encoder.
	_ "zettelstore.de/z/encoder/jsonenc"   // Allow to use JSON encoder.
	_ "zettelstore.de/z/encoder/nativeenc" // Allow to use native encoder.
	_ "zettelstore.de/z/encoder/rawenc"    // Allow to use raw encoder.
	_ "zettelstore.de/z/encoder/textenc"   // Allow to use text encoder.
	_ "zettelstore.de/z/encoder/zmkenc"    // Allow to use zmk encoder.
	_ "zettelstore.de/z/parser/blob"       // Allow to use BLOB parser.
	_ "zettelstore.de/z/parser/markdown"   // Allow to use markdown parser.
	_ "zettelstore.de/z/parser/meta"       // Allow to use meta parser.
	_ "zettelstore.de/z/parser/plain"      // Allow to use plain parser.
	_ "zettelstore.de/z/parser/zettelmark" // Allow to use zettelmark parser.
	_ "zettelstore.de/z/place/constplace"  // Allow to use global internal place.
	_ "zettelstore.de/z/place/dirplace"    // Allow to use directory place.
	_ "zettelstore.de/z/place/memplace"    // Allow to use memory place.
)
