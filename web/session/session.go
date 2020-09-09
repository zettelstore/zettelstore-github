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

// Package session provides utilities for using sessions.
package session

import (
	"context"
	"net/http"

	"zettelstore.de/z/auth"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

const sessionName = "zsession"

// SetToken sets the session cookie for later user identification.
func SetToken(w http.ResponseWriter, token []byte) {
	cookie := http.Cookie{
		Name:     sessionName,
		Value:    string(token),
		Path:     config.URLPrefix(),
		Secure:   config.SecureCookie(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

// ClearToken invalidates the session cookie by sending an empty one.
func ClearToken(ctx context.Context, w http.ResponseWriter) context.Context {
	SetToken(w, nil)
	return updateContext(ctx, nil)
}

// Handler enriches the request context with optional user information.
type Handler struct {
	next         http.Handler
	getUserByZid usecase.GetUserByZid
}

// NewHandler creates a new handler.
func NewHandler(next http.Handler, getUserByZid usecase.GetUserByZid) *Handler {
	return &Handler{
		next:         next,
		getUserByZid: getUserByZid,
	}
}

type contextUser struct{}

var contextKey contextUser

// GetUser returns the user meta data from the context, if there is one. Else return nil.
func GetUser(ctx context.Context) *domain.Meta {
	user, ok := ctx.Value(contextKey).(*domain.Meta)
	if ok {
		return user
	}
	return nil
}

func updateContext(ctx context.Context, user *domain.Meta) context.Context {
	return context.WithValue(ctx, contextKey, user)
}

// ServeHTTP processes one HTTP request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		h.next.ServeHTTP(w, r)
		return
	}
	token := []byte(cookie.Value)
	ident, zid, err := auth.CheckToken(token)
	if err != nil {
		h.next.ServeHTTP(w, r)
		return
	}
	ctx := r.Context()
	user, err := h.getUserByZid.Run(ctx, zid, ident)
	if err != nil {
		h.next.ServeHTTP(w, r)
		return
	}
	h.next.ServeHTTP(w, r.WithContext(updateContext(ctx, user)))
}
