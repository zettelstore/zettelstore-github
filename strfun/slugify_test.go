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
package strfun_test

import (
	"testing"

	"zettelstore.de/z/strfun"
)

var tests = []struct{ in, exp string }{
	{"simple test", "simple-test"},
	{"I'm a go developer", "i-m-a-go-developer"},
	{"-!->simple   test<-!-", "simple-test"},
	{"äöüÄÖÜß", "aouaouß"},
	{"\"aèf", "aef"},
	{"a#b", "a-b"},
}

func TestSlugify(t *testing.T) {
	for _, test := range tests {
		if got := strfun.Slugify(test.in); got != test.exp {
			t.Errorf("%q: %q != %q", test.in, got, test.exp)
		}
	}
}
