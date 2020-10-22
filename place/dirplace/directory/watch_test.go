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

// Package directory manages the directory part of a directory place.
package directory

import (
	"testing"
)

func sameStringSlices(sl1, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i := 0; i < len(sl1); i++ {
		if sl1[i] != sl2[i] {
			return false
		}
	}
	return true
}

func TestMatchValidFileName(t *testing.T) {
	testcases := []struct {
		name string
		exp  []string
	}{
		{"", []string{}},
		{".txt", []string{}},
		{"12345678901234.txt", []string{"12345678901234", ".txt", "txt"}},
		{"12345678901234abc.txt", []string{"12345678901234", ".txt", "txt"}},
		{"12345678901234.abc.txt", []string{"12345678901234", ".txt", "txt"}},
	}

	for i, tc := range testcases {
		got := matchValidFileName(tc.name)
		if len(got) == 0 {
			if len(tc.exp) > 0 {
				t.Errorf("TC=%d, name=%q, exp=%v, got=%v", i, tc.name, tc.exp, got)
			}
		} else {
			if got[0] != tc.name {
				t.Errorf("TC=%d, name=%q, got=%v", i, tc.name, got)
			}
			if !sameStringSlices(got[1:], tc.exp) {
				t.Errorf("TC=%d, name=%q, exp=%v, got=%v", i, tc.name, tc.exp, got)
			}
		}
	}
}
