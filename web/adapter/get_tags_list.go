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
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"zettelstore.de/z/encoder"
	"zettelstore.de/z/encoder/jsonenc"
	"zettelstore.de/z/usecase"
)

// MakeListTagsHandler creates a new HTTP handler for the use case "list some zettel".
func MakeListTagsHandler(te *TemplateEngine, listTags usecase.ListTags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iMinCount, _ := strconv.Atoi(r.URL.Query().Get("min"))
		tagData, err := listTags.Run(ctx, iMinCount)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		format := getFormat(r, r.URL.Query(), encoder.GetDefaultFormat())
		switch format {
		case "json":
			w.Header().Set("Content-Type", format2ContentType(format))
			renderListTagsJSON(w, tagData)
		default:
			http.Error(w, fmt.Sprintf("Tags list not available in format %q", format), http.StatusBadRequest)
		}
	}
}

func renderListTagsJSON(w http.ResponseWriter, tagData usecase.TagData) {
	buf := encoder.NewBufWriter(w)

	tagList := make([]string, 0, len(tagData))
	for tag := range tagData {
		tagList = append(tagList, tag)
	}
	sort.Strings(tagList)

	buf.WriteString("{\"tags\":{")
	first := true
	for _, tag := range tagList {
		if first {
			buf.WriteByte('"')
			first = false
		} else {
			buf.WriteString(",\"")
		}
		buf.Write(jsonenc.Escape(tag))
		buf.WriteString("\":[")
		for i, meta := range tagData[tag] {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('"')
			buf.WriteString(meta.Zid.Format())
			buf.WriteByte('"')
		}
		buf.WriteString("]")

	}
	buf.WriteString("}}")
	buf.Flush()
}
