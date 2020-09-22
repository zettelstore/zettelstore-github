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
	"encoding/json"
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
		if !config.WithAuth() {
			http.Error(w, "Authentication not available", http.StatusForbidden)
			return
		}
		htmlDur, apiDur := config.TokenLifetime()
		switch format := getFormat(r, "html"); format {
		case "html":
			authenticateViaHTML(te, auth, w, r, htmlDur)
		case "json":
			authenticateViaJSON(auth, w, r, apiDur)
		default:
			http.Error(w, fmt.Sprintf("Authentication not available in format %q", format), http.StatusBadRequest)
		}
	}
}

func authenticateViaHTML(te *TemplateEngine, auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) {
	ident, cred, ok := getCredentialsViaForm(r)
	if !ok {
		http.Error(w, "Unable to read login form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	token, err := auth.Run(ctx, ident, cred, authDuration)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}
	if token == nil {
		renderLoginForm(session.ClearToken(ctx, w), w, te, true)
		return
	}

	session.SetToken(w, token, authDuration)
	http.Redirect(w, r, urlForList('/'), http.StatusFound)
}

func authenticateViaJSON(auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) {
	token, err := authenticateForJSON(auth, w, r, authDuration)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}
	if token == nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", `Bearer realm="Default"`)
		return
	}

	w.Header().Set("Content-Type", format2ContentType("json"))
	je := json.NewEncoder(w)
	je.Encode(struct {
		Token string `json:"token"`
	}{
		Token: string(token),
	})
}

func authenticateForJSON(auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) ([]byte, error) {
	ident, cred, ok := getCredentialsViaForm(r)
	if !ok {
		if ident, cred, ok = getCredentialsViaBasicAuth(r); !ok {
			return nil, nil
		}
	}
	token, err := auth.Run(r.Context(), ident, cred, authDuration)
	return token, err
}

func getCredentialsViaForm(r *http.Request) (ident, cred string, ok bool) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return "", "", false
	}

	ident = strings.TrimSpace(r.PostFormValue("username"))
	cred = r.PostFormValue("password")
	if ident == "" {
		return "", "", false
	}
	return ident, cred, true
}

func getCredentialsViaBasicAuth(r *http.Request) (ident, cred string, ok bool) {
	return r.BasicAuth()
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
