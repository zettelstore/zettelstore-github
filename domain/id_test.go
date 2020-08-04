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

// Package domain_test provides unit tests for testing domain specific functions.
package domain_test

import (
	"testing"

	"zettelstore.de/z/domain"
)

func TestIsValid(t *testing.T) {
	validIDs := []string{
		"00000000000000",
		"99999999999999",
		"20200310195100",
	}

	for i, id := range validIDs {
		if !domain.ZettelID(id).IsValid() || !domain.IsValidID(id) {
			t.Errorf("i=%d: id=%q is not valid, but should be", i, id)
		}
	}

	invalidIDs := []string{
		"", "0", "a",
		"000000000000000",
		"99999999999999a",
		"20200310T195100",
	}

	for i, id := range invalidIDs {
		if domain.ZettelID(id).IsValid() || domain.IsValidID(id) {
			t.Errorf("i=%d: id=%q is valid, but should not be", i, id)
		}
	}
}
