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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"zettelstore.de/z/auth/token"
	"zettelstore.de/z/config"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakePostLoginHandler creates a new HTTP handler to authenticate the given user.
func MakePostLoginHandler(te *TemplateEngine, auth usecase.Authenticate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := getFormat(r, r.URL.Query(), encoder.GetDefaultFormat())
		if !config.WithAuth() {
			switch format {
			case "html":
				http.Redirect(w, r, newURLBuilder('/').String(), http.StatusFound)
			case "json":
				w.Header().Set("Content-Type", format2ContentType("json"))
				writeJSONToken(w, "freeaccess", 24*366*10*time.Hour)
			default:
				http.Error(w, "Unknown format", http.StatusBadRequest)
			}
			return
		}
		htmlDur, apiDur := config.TokenLifetime()
		switch format {
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
	token, err := auth.Run(ctx, ident, cred, authDuration, token.KindHTML)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}
	if token == nil {
		renderLoginForm(session.ClearToken(ctx, w), w, te, true)
		return
	}

	session.SetToken(w, token, authDuration)
	http.Redirect(w, r, newURLBuilder('/').String(), http.StatusFound)
}

func authenticateViaJSON(auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) {
	token, err := authenticateForJSON(auth, w, r, authDuration)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}
	if token == nil {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Default"`)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", format2ContentType("json"))
	writeJSONToken(w, string(token), authDuration)
}

func authenticateForJSON(auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) ([]byte, error) {
	ident, cred, ok := getCredentialsViaForm(r)
	if !ok {
		if ident, cred, ok = getCredentialsViaBasicAuth(r); !ok {
			return nil, nil
		}
	}
	token, err := auth.Run(r.Context(), ident, cred, authDuration, token.KindJSON)
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

func writeJSONToken(w http.ResponseWriter, token string, lifetime time.Duration) {
	je := json.NewEncoder(w)
	je.Encode(struct {
		Token   string `json:"access_token"`
		Type    string `json:"token_type"`
		Expires int    `json:"expires_in"`
	}{
		Token:   token,
		Type:    "Bearer",
		Expires: int(lifetime / time.Second),
	})
}

// MakeGetLogoutHandler creates a new HTTP handler to log out the current user
func MakeGetLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := getFormat(r, r.URL.Query(), "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Logout not possible in format %q", format), http.StatusBadRequest)
			return
		}

		session.ClearToken(r.Context(), w)
		http.Redirect(w, r, newURLBuilder('/').String(), http.StatusFound)
	}
}

// MakeRenewAuthHandler creates a new HTTP handler to renew the authenticate of a user.
func MakeRenewAuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := session.GetAuthData(ctx)
		if auth == nil || auth.Token == nil || auth.User == nil {
			http.Error(w, "Not authenticated", http.StatusBadRequest)
			return
		}
		totalLifetime := auth.Expires.Sub(auth.Issued)
		currentLifetime := auth.Now.Sub(auth.Issued)
		// If we are in the first quarter of the tokens lifetime, return the token
		if currentLifetime*4 < totalLifetime {
			w.Header().Set("Content-Type", format2ContentType("json"))
			writeJSONToken(w, string(auth.Token), totalLifetime-currentLifetime)
			return
		}

		// Toke is a little bit aged. Create a new one
		_, apiDur := config.TokenLifetime()
		token, err := token.GetToken(auth.User, apiDur, token.KindJSON)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}
		w.Header().Set("Content-Type", format2ContentType("json"))
		writeJSONToken(w, string(token), apiDur)
	}
}
