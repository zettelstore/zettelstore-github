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

// Package auth provides some function for authentication.
package auth

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/pascaldekloe/jwt"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

// HashCredential returns a hashed vesion of the given credential
func HashCredential(credential string) (string, error) {
	res, err := bcrypt.GenerateFromPassword([]byte(credential), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// CompareHashAndCredential checks, whether the hashedCredential is a possible
// value when hashing the credential.
func CompareHashAndCredential(hashedCredential string, credential string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCredential), []byte(credential))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, err
}

const reqHash = jwt.HS512

// ErrNoUser signals that the meta data has no role value 'user'.
var ErrNoUser = errors.New("auth: meta is no user")

// ErrNoIdent signals that the 'ident' key is missing.
var ErrNoIdent = errors.New("auth: missing ident")

// ErrNoZid signals that the 'zid' key is missing.
var ErrNoZid = errors.New("auth: missing zettel id")

// GetToken returns a token to be used for authentification
func GetToken(ident *domain.Meta, d time.Duration) ([]byte, error) {
	if role, ok := ident.Get(domain.MetaKeyRole); !ok || role != "user" {
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
		},
	}
	token, err := claims.HMACSign(reqHash, config.GetSecret())
	if err != nil {
		return nil, err
	}
	return token, nil
}

// ErrTokenExpired signals an exired token
var ErrTokenExpired = errors.New("auth: token expired")

// CheckToken checks the validity of the token and returns relevant data.
func CheckToken(token []byte) (string, domain.ZettelID, error) {
	h, err := jwt.NewHMAC(reqHash, config.GetSecret())
	if err != nil {
		return "", domain.InvalidZettelID, err
	}
	claims, err := h.Check(token)
	if err != nil {
		return "", domain.InvalidZettelID, err
	}
	now := time.Now().Round(time.Second)
	if claims.Expires.Time().Before(now) {
		return "", domain.InvalidZettelID, ErrTokenExpired
	}
	ident := claims.Subject
	if len(ident) == 0 {
		return "", domain.InvalidZettelID, ErrNoIdent
	}
	if zidS, ok := claims.Set["zid"].(string); ok {
		if zid, err := domain.ParseZettelID(zidS); err == nil {
			return ident, zid, nil
		}
	}
	return "", domain.InvalidZettelID, ErrNoZid
}
