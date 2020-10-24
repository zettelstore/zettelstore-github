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

package domain_test

import (
	"testing"

	"zettelstore.de/z/domain"
)

func TestContentIsBinary(t *testing.T) {
	td := []struct {
		s   string
		exp bool
	}{
		{"abc", true},
		{"äöü", true},
		{"", true},
		{string([]byte{0}), false},
	}
	for i, tc := range td {
		content := domain.NewContent(tc.s)
		got := content.IsBinary()
		if got != tc.exp {
			t.Errorf("TC=%d: expected %v, got %v", i, tc.exp, got)
		}
	}
}
