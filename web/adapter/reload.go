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
	"net/http"

	"zettelstore.de/z/encoder"
	"zettelstore.de/z/usecase"
)

// MakeReloadHandler creates a new HTTP handler for the use case "reload".
func MakeReloadHandler(reload usecase.Reload) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := reload.Run(r.Context())
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		format := getFormat(r, r.URL.Query(), encoder.GetDefaultFormat())
		if format == "html" {
			http.Redirect(w, r, newURLBuilder('/').String(), http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", format2ContentType(format))
		w.WriteHeader(http.StatusNoContent)
	}
}
