//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package webui provides wet-UI handlers for web requests.
package webui

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zettelstore.de/z/auth/token"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/session"
)

// MakeGetLoginHandler creates a new HTTP handler to display the HTML login view.
func MakeGetLoginHandler(te *TemplateEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderLoginForm(session.ClearToken(r.Context(), w), w, te, false)
	}
}

func renderLoginForm(ctx context.Context, w http.ResponseWriter, te *TemplateEngine, retry bool) {
	te.renderTemplate(ctx, w, domain.LoginTemplateID, struct {
		baseData
		Retry bool
	}{
		baseData: te.makeBaseData(ctx, config.GetDefaultLang(), "Login", nil),
		Retry:    retry,
	})
}

// MakePostLoginHandlerHTML creates a new HTTP handler to authenticate the given user.
func MakePostLoginHandlerHTML(te *TemplateEngine, auth usecase.Authenticate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !config.WithAuth() {
			http.Redirect(w, r, adapter.NewURLBuilder('/').String(), http.StatusFound)
			return
		}
		htmlDur, _ := config.TokenLifetime()
		authenticateViaHTML(te, auth, w, r, htmlDur)
	}
}

func authenticateViaHTML(te *TemplateEngine, auth usecase.Authenticate, w http.ResponseWriter, r *http.Request, authDuration time.Duration) {
	ident, cred, ok := adapter.GetCredentialsViaForm(r)
	if !ok {
		http.Error(w, "Unable to read login form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	token, err := auth.Run(ctx, ident, cred, authDuration, token.KindHTML)
	if err != nil {
		adapter.ReportUsecaseError(w, err)
		return
	}
	if token == nil {
		renderLoginForm(session.ClearToken(ctx, w), w, te, true)
		return
	}

	session.SetToken(w, token, authDuration)
	http.Redirect(w, r, adapter.NewURLBuilder('/').String(), http.StatusFound)
}

// MakeGetLogoutHandler creates a new HTTP handler to log out the current user
func MakeGetLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if format := adapter.GetFormat(r, r.URL.Query(), "html"); format != "html" {
			http.Error(w, fmt.Sprintf("Logout not possible in format %q", format), http.StatusBadRequest)
			return
		}

		session.ClearToken(r.Context(), w)
		http.Redirect(w, r, adapter.NewURLBuilder('/').String(), http.StatusFound)
	}
}
