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
	"log"
	"net/http"

	"zettelstore.de/z/place"
)

// ReportUsecaseError returns an appropriate HTTP status code for errors in use cases.
func ReportUsecaseError(w http.ResponseWriter, err error) {
	if err, ok := err.(*place.ErrUnknownID); ok {
		http.Error(w, fmt.Sprintf("Zettel %q not found", err.Zid.Format()), http.StatusNotFound)
		return
	}
	if err, ok := err.(*place.ErrNotAllowed); ok {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err, ok := err.(*place.ErrInvalidID); ok {
		http.Error(w, fmt.Sprintf("Zettel-ID %q not appropriate in this context", err.Zid.Format()), http.StatusBadRequest)
		return
	}
	if err == place.ErrStopped {
		http.Error(w, "Zettelstore not operational", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Unknown internal error", http.StatusInternalServerError)
	log.Println(err)
}
