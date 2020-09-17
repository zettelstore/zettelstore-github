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
	"strings"
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
			http.Error(w, fmt.Sprintf("Login not possible in format %q", format), http.StatusBadRequest)
			return
		}

		if !config.WithAuth() {
			http.Error(w, "Login not available", http.StatusForbidden)
			return
		}

		renderLoginForm(session.ClearToken(r.Context(), w), w, te, false)
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
		User:  wrapUser(nil),
		Retry: retry,
	})
}

// MakePostLoginHandler creates a new HTTP handler to authenticate the given user.
func MakePostLoginHandler(te *TemplateEngine, auth usecase.Authenticate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		htmlDur, apiDur := config.TokenLifetime()
		var formatDur time.Duration
		var formatCode int
		switch format := getFormat(r, "html"); format {
		case "html":
			formatCode = 1
			formatDur = htmlDur
		case "json":
			formatCode = 2
			formatDur = apiDur
		default:
			http.Error(w, fmt.Sprintf("Authentication not available in format %q", format), http.StatusBadRequest)
			return
		}
		if !config.WithAuth() {
			http.Error(w, "Authentication not available", http.StatusForbidden)
			return
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to read login form", http.StatusBadRequest)
			log.Println(err)
			return
		}

		ident := strings.TrimSpace(r.PostFormValue("username"))
		cred := r.PostFormValue("password")
		ctx := r.Context()
		token, err := auth.Run(ctx, ident, cred, formatDur)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}
		if token == nil {
			switch formatCode {
			case 1:
				renderLoginForm(session.ClearToken(ctx, w), w, te, true)
			case 2:
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				w.Header().Set("WWW-Authenticate", `Bearer realm="Default"`)
			}
			return
		}

		switch formatCode {
		case 1:
			session.SetToken(w, token, formatDur)
			http.Redirect(w, r, urlForList('/'), http.StatusFound)
		case 2:
		}
	}
}

// MakeGetLogoutHandler creates a new HTTP handler to log out the current user
func MakeGetLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Logout not possible in format %q", format), http.StatusBadRequest)
			return
		}

		session.ClearToken(r.Context(), w)
		http.Redirect(w, r, urlForList('/'), http.StatusFound)
	}
}
