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
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakeGetLoginHandler creates a new HTTP handler to display the HTML login view.
func MakeGetLoginHandler(te *TemplateEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Login not possible in format %q", format), http.StatusNotFound)
			return
		}

		if !config.GetOwner().IsValid() {
			http.Error(w, "Login not available", http.StatusBadRequest)
			return
		}

		renderLoginForm(r.Context(), w, te, false)
	}
}

func renderLoginForm(ctx context.Context, w http.ResponseWriter, te *TemplateEngine, retry bool) {
	te.renderTemplate(ctx, w, domain.LoginTemplateID, struct {
		Lang  string
		Title string
		User  userWrapper
		Retry bool
	}{
		Lang:  config.GetDefaultLang(),
		Title: "Login",
		User:  wrapUser(session.GetUser(ctx)),
		Retry: retry,
	})
}

// MakePostLoginHandler creates a new HTTP handler to authenticate the given user.
func MakePostLoginHandler(te *TemplateEngine, auth usecase.Authenticate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := getFormat(r, "html")
		if !config.GetOwner().IsValid() {
			http.Error(w, "Authentication not available", http.StatusBadRequest)
			return
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to read login form", http.StatusBadRequest)
			log.Println(err)
			return
		}

		ident := r.PostFormValue("username")
		cred := r.PostFormValue("password")
		d := time.Second * 600 // TODO: longer for HTML, configurable, ...
		ctx := r.Context()
		token, err := auth.Run(ctx, ident, cred, d)
		if err != nil {
			http.Error(w, "Unable to check login data", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if token == nil {
			switch format {
			case "html":
				renderLoginForm(session.ClearToken(ctx, w), w, te, true)
			default:
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				w.Header().Set("WWW-Authenticate", `Bearer realm="Default"`)
			}
			return
		}

		switch format {
		case "html":
			session.SetToken(w, token)
			http.Redirect(w, r, urlForList('/'), http.StatusFound)
		default:
		}
	}
}

// MakeGetLogoutHandler creates a new HTTP handler to log out the current user
func MakeGetLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Logout not possible in format %q", format), http.StatusNotFound)
			return
		}

		session.ClearToken(r.Context(), w)
		http.Redirect(w, r, urlForList('/'), http.StatusFound)
	}
}
