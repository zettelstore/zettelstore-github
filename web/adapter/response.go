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
	"log"
	"net/http"

	"zettelstore.de/z/place"
)

func checkUsecaseError(w http.ResponseWriter, err error) {
	if err, ok := err.(*place.ErrUnknownID); ok {
		http.Error(w, fmt.Sprintf("Zettel %q not found", err.Zid.Format()), http.StatusNotFound)
		return
	}
	if err, ok := err.(*place.ErrNotAuthorized); ok {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
