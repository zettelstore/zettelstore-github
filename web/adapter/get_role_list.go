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

	"zettelstore.de/z/encoder"
	"zettelstore.de/z/encoder/jsonenc"
	"zettelstore.de/z/usecase"
)

// MakeListRoleHandler creates a new HTTP handler for the use case "list some zettel".
func MakeListRoleHandler(te *TemplateEngine, listRole usecase.ListRole) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		roleList, err := listRole.Run(ctx)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		format := getFormat(r, encoder.GetDefaultFormat())
		switch format {
		case "json":
			w.Header().Set("Content-Type", format2ContentType(format))
			renderListRoleJSON(w, roleList)
		default:
			http.Error(w, fmt.Sprintf("Role list not available in format %q", format), http.StatusBadRequest)
		}

	}
}

func renderListRoleJSON(w http.ResponseWriter, roleList []string) {
	buf := encoder.NewBufWriter(w)

	buf.WriteString("{\"role-list\":[")
	first := true
	for _, role := range roleList {
		if first {
			buf.WriteByte('"')
			first = false
		} else {
			buf.WriteString("\",\"")
		}
		buf.Write(jsonenc.Escape(role))
	}
	if !first {
		buf.WriteByte('"')
	}
	buf.WriteString("]}")
	buf.Flush()
}
