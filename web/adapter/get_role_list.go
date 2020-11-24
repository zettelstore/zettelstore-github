//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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

		format := getFormat(r, r.URL.Query(), encoder.GetDefaultFormat())
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
