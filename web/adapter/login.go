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

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

type loginData struct {
	Lang  string
	Title string
}

// MakeGetLoginHandler creates a new HTTP handler to display the HTML login view.
func MakeGetLoginHandler(te *TemplateEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Login not possible in format %q", format), http.StatusNotFound)
			return
		}

		te.renderTemplate(r.Context(), w, domain.LoginTemplateID, loginData{
			Lang:  config.Config.GetDefaultLang(),
			Title: "Login",
		})
	}
}

// MakePostLoginHandler creates a new HTTP handler to authenticate the given user.
func MakePostLoginHandler(auth usecase.Authenticate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to read login form", http.StatusBadRequest)
			log.Println(err)
			return
		}

		ident := r.PostFormValue("username")
		cred := r.PostFormValue("password")
		ok, err := auth.Run(r.Context(), ident, cred)
		if err != nil {
			http.Error(w, "Unable to check login data", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if ok {
			successfulLogin(w, r, ident)
		} else {
			failedLogin(w, r)
		}
	}
}

func successfulLogin(w http.ResponseWriter, r *http.Request, ident string) {
	switch format := getFormat(r, "html"); format {
	case "html":
		http.Redirect(w, r, urlFor('/', ""), http.StatusFound)
	default:
	}
}

func failedLogin(w http.ResponseWriter, r *http.Request) {
	switch format := getFormat(r, "html"); format {
	case "html":
		http.Redirect(w, r, urlFor('a', ""), http.StatusFound)
	default:
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", `Bearer realm="Default"`)
	}
}
