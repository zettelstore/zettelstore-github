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

// Package input_test provides some unit-tests for reading data.
package input_test

import (
	"testing"

	"zettelstore.de/z/input"
)

func TestEatEOL(t *testing.T) {
	inp := input.NewInput("")
	inp.EatEOL()
	if inp.Ch != input.EOS {
		t.Errorf("No EOS found: %q", inp.Ch)
	}
	if inp.Pos != 0 {
		t.Errorf("Pos != 0: %d", inp.Pos)
	}

	inp = input.NewInput("ABC")
	if inp.Ch != 'A' {
		t.Errorf("First ch != 'A', got %q", inp.Ch)
	}
	inp.EatEOL()
	if inp.Ch != 'A' {
		t.Errorf("First ch != 'A', got %q", inp.Ch)
	}
}

func TestScanEntity(t *testing.T) {
	var testcases = []struct {
		text string
		exp  string
	}{
		{"", ""},
		{"a", ""},
		{"&amp;", "&"},
		{"&#9;", "\t"},
		{"&quot;", "\""},
	}
	for id, tc := range testcases {
		inp := input.NewInput(tc.text)
		got, ok := inp.ScanEntity()
		if !ok {
			if tc.exp != "" {
				t.Errorf("ID=%d, text=%q: expected error, but got %q", id, tc.text, got)
			}
			if inp.Pos != 0 {
				t.Errorf("ID=%d, text=%q: input position advances to %d", id, tc.text, inp.Pos)
			}
			continue
		}
		if tc.exp != got {
			t.Errorf("ID=%d, text=%q: expected %q, but got %q", id, tc.text, tc.exp, got)
		}
	}
}
