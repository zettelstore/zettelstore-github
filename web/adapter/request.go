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
	"zettelstore.de/z/store"
)

func getFormat(r *http.Request, defFormat string) string {
	format := r.URL.Query().Get("_format")
	if len(format) > 0 {
		return format
	}
	if accepts, ok := r.Header["Accept"]; ok {
		for _, acc := range accepts {
			switch acc {
			case "application/json":
				return "json"
			case "text/html":
				return "html"
			}
		}
	}
	return defFormat
}

var formatCT = map[string]string{
	"html":   "text/html; charset=utf-8",
	"native": "text/plain; charset=utf-8",
	"json":   "application/json",
	"djson":  "application/json",
	"text":   "text/plain; charset=utf-8",
	"zmk":    "text/plain; charset=utf-8",
}

func formatContentType(format string) string {
	ct, ok := formatCT[format]
	if !ok {
		return "application/octet-stream"
	}
	return ct
}

func getFilterSorter(r *http.Request) (filter *store.Filter, sorter *store.Sorter) {
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

func ensureFilter(filter *store.Filter) *store.Filter {
	if filter == nil {
		filter = new(store.Filter)
		filter.Expr = make(store.FilterExpr)
	}
	return filter
}

func ensureSorter(sorter *store.Sorter) *store.Sorter {
	if sorter == nil {
		sorter = new(store.Sorter)
	}
	return sorter
}
