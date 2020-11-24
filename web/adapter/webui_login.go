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
	"context"
	"net/http"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
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
