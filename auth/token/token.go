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

// Package token provides some function for handling auth token.
package token

import (
	"errors"
	"time"

	"github.com/pascaldekloe/jwt"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

const reqHash = jwt.HS512

// ErrNoUser signals that the meta data has no role value 'user'.
var ErrNoUser = errors.New("auth: meta is no user")

// ErrNoIdent signals that the 'ident' key is missing.
var ErrNoIdent = errors.New("auth: missing ident")

// ErrOtherKind signals that the token was defined for another token kind.
var ErrOtherKind = errors.New("auth: wrong token kind")

// ErrNoZid signals that the 'zid' key is missing.
var ErrNoZid = errors.New("auth: missing zettel id")

// Kind specifies for which application / usage a token is/was requested.
type Kind int

// Allowed values of token kind
const (
	_ Kind = iota
	KindJSON
	KindHTML
)

// GetToken returns a token to be used for authentification.
func GetToken(ident *domain.Meta, d time.Duration, kind Kind) ([]byte, error) {
	if role, ok := ident.Get(domain.MetaKeyRole); !ok || role != domain.MetaValueRoleUser {
		return nil, ErrNoUser
	}
	subject, ok := ident.Get(domain.MetaKeyIdent)
	if !ok || len(subject) == 0 {
		return nil, ErrNoIdent
	}

	now := time.Now().Round(time.Second)
	claims := jwt.Claims{
		Registered: jwt.Registered{
			Subject: subject,
			Expires: jwt.NewNumericTime(now.Add(d)),
			Issued:  jwt.NewNumericTime(now),
		},
		Set: map[string]interface{}{
			"zid": ident.Zid.Format(),
			"_tk": int(kind),
		},
	}
	token, err := claims.HMACSign(reqHash, config.Secret())
	if err != nil {
		return nil, err
	}
	return token, nil
}

// ErrTokenExpired signals an exired token
var ErrTokenExpired = errors.New("auth: token expired")

// Data contains some important elements from a token.
type Data struct {
	Token   []byte
	Now     time.Time
	Issued  time.Time
	Expires time.Time
	Ident   string
	Zid     domain.ZettelID
}

// CheckToken checks the validity of the token and returns relevant data.
func CheckToken(token []byte, k Kind) (Data, error) {
	h, err := jwt.NewHMAC(reqHash, config.Secret())
	if err != nil {
		return Data{}, err
	}
	claims, err := h.Check(token)
	if err != nil {
		return Data{}, err
	}
	now := time.Now().Round(time.Second)
	expires := claims.Expires.Time()
	if expires.Before(now) {
		return Data{}, ErrTokenExpired
	}
	ident := claims.Subject
	if len(ident) == 0 {
		return Data{}, ErrNoIdent
	}
	if zidS, ok := claims.Set["zid"].(string); ok {
		if zid, err := domain.ParseZettelID(zidS); err == nil {
			if kind, ok := claims.Set["_tk"].(float64); ok {
				if Kind(kind) == k {
					return Data{
						Token:   token,
						Now:     now,
						Issued:  claims.Issued.Time(),
						Expires: expires,
						Ident:   ident,
						Zid:     zid,
					}, nil
				}
			}
			return Data{}, ErrOtherKind
		}
	}
	return Data{}, ErrNoZid
}
