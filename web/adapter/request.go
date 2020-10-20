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

// Package adapter provides handlers for web requests.
package adapter

import (
	"net/http"
	"strconv"
	"strings"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
)

func getFormat(r *http.Request, defFormat string) string {
	format := r.URL.Query().Get("_format")
	if len(format) > 0 {
		return format
	}
	if format, ok := getOneFormat(r, "Accept"); ok {
		return format
	}
	if format, ok := getOneFormat(r, "Content-Type"); ok {
		return format
	}
	return defFormat
}

func getOneFormat(r *http.Request, key string) (string, bool) {
	if values, ok := r.Header[key]; ok {
		for _, value := range values {
			if format, ok := contentType2format(value); ok {
				return format, true
			}
		}
	}
	return "", false
}

func getPart(r *http.Request, defPart string) string {
	part := r.URL.Query().Get("_part")
	if len(part) == 0 {
		part = defPart
	}
	return part
}

func getFilterSorter(r *http.Request) (filter *place.Filter, sorter *place.Sorter) {
	for key, values := range r.URL.Query() {
		switch key {
		case "_sort":
			if len(values) > 0 {
				descending := false
				sortkey := values[0]
				if strings.HasPrefix(sortkey, "-") {
					descending = true
					sortkey = sortkey[1:]
				}
				if domain.KeyIsValid(sortkey) {
					sorter = ensureSorter(sorter)
					sorter.Order = sortkey
					sorter.Descending = descending
				}
			}
		case "_offset":
			if len(values) > 0 {
				if offset, err := strconv.Atoi(values[0]); err == nil {
					sorter = ensureSorter(sorter)
					sorter.Offset = offset
				}
			}
		case "_limit":
			if len(values) > 0 {
				if limit, err := strconv.Atoi(values[0]); err == nil {
					sorter = ensureSorter(sorter)
					sorter.Limit = limit
				}
			}
		case "_negate":
			filter = ensureFilter(filter)
			filter.Negate = true
		case "_s":
			cleanedValues := make([]string, 0, len(values))
			for _, val := range values {
				if len(val) > 0 {
					cleanedValues = append(cleanedValues, val)
				}
			}
			if len(cleanedValues) > 0 {
				filter = ensureFilter(filter)
				filter.Expr[""] = cleanedValues
			}
		default:
			if domain.KeyIsValid(key) {
				filter = ensureFilter(filter)
				filter.Expr[key] = values
			}
		}
	}
	return filter, sorter
}

func ensureFilter(filter *place.Filter) *place.Filter {
	if filter == nil {
		filter = new(place.Filter)
		filter.Expr = make(place.FilterExpr)
	}
	return filter
}

func ensureSorter(sorter *place.Sorter) *place.Sorter {
	if sorter == nil {
		sorter = new(place.Sorter)
	}
	return sorter
}
