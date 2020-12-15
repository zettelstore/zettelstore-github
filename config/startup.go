//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package config provides functions to retrieve configuration data.
package config

import (
	"hash/fnv"
	"strconv"
	"time"

	"zettelstore.de/z/domain"
)

var startupConfig struct {
	readonlyMode  bool
	urlPrefix     string
	insecCookie   bool
	persistCookie bool
	owner         domain.ZettelID
	withAuth      bool
	secret        []byte
	htmlLifetime  time.Duration
	apiLifetime   time.Duration
}

// Predefined keys for startup zettel
const (
	StartupKeyInsecureCookie    = "insecure-cookie"
	StartupKeyListenAddress     = "listen-addr"
	StartupKeyOwner             = "owner"
	StartupKeyPersistentCookie  = "persistent-cookie"
	StartupKeyPlaceOneURI       = "place-1-uri"
	StartupKeyReadOnlyMode      = "read-only-mode"
	StartupKeyTargetFormat      = "target-format"
	StartupKeyTokenLifetimeHTML = "token-lifetime-html"
	StartupKeyTokenLifetimeAPI  = "token-lifetime-api"
	StartupKeyURLPrefix         = "url-prefix"
	StartupKeyVerbose           = "verbose"
)

// SetupStartup initializes the startup data.
func SetupStartup(cfg *domain.Meta) {
	if startupConfig.urlPrefix != "" {
		panic("startupConfig already set")
	}
	startupConfig.readonlyMode = cfg.GetBool(StartupKeyReadOnlyMode)
	startupConfig.urlPrefix = cfg.GetDefault(StartupKeyURLPrefix, "/")
	startupConfig.owner = domain.InvalidZettelID
	if owner, ok := cfg.Get(StartupKeyOwner); ok {
		if zid, err := domain.ParseZettelID(owner); err == nil {
			startupConfig.owner = zid
			startupConfig.withAuth = true
		}
	}
	if startupConfig.withAuth {
		startupConfig.insecCookie = cfg.GetBool(StartupKeyInsecureCookie)
		startupConfig.persistCookie = cfg.GetBool(StartupKeyPersistentCookie)
		startupConfig.secret = calcSecret(cfg)
		startupConfig.htmlLifetime = getDuration(
			cfg, StartupKeyTokenLifetimeHTML, 1*time.Hour, 1*time.Minute, 30*24*time.Hour)
		startupConfig.apiLifetime = getDuration(
			cfg, StartupKeyTokenLifetimeAPI, 10*time.Minute, 0, 1*time.Hour)
	}
}

func calcSecret(cfg *domain.Meta) []byte {
	h := fnv.New128()
	if secret, ok := cfg.Get("secret"); ok {
		h.Write([]byte(secret))
	}
	h.Write([]byte(version.Prog))
	h.Write([]byte(version.Build))
	h.Write([]byte(version.Hostname))
	h.Write([]byte(version.GoVersion))
	h.Write([]byte(version.Os))
	h.Write([]byte(version.Arch))
	return h.Sum(nil)
}

func getDuration(cfg *domain.Meta, key string, defDur, minDur, maxDur time.Duration) time.Duration {
	if s, ok := cfg.Get(key); ok && len(s) > 0 {
		if d, err := strconv.ParseUint(s, 10, 64); err == nil {
			secs := time.Duration(d) * time.Minute
			if secs < minDur {
				return minDur
			}
			if secs > maxDur {
				return maxDur
			}
			return secs
		}
	}
	return defDur
}

// IsReadOnlyMode returns whether the system is in read-only mode or not.
func IsReadOnlyMode() bool { return startupConfig.readonlyMode }

// URLPrefix returns the configured prefix to be used when providing URL to
// the service.
func URLPrefix() string { return startupConfig.urlPrefix }

// SecureCookie returns whether the web app should set cookies to secure mode.
func SecureCookie() bool { return !startupConfig.insecCookie }

// PersistentCookie returns whether the web app should set persistent cookies
// (instead of temporary).
func PersistentCookie() bool { return startupConfig.persistCookie }

// Owner returns the zid of the zettelkasten's owner.
// If there is no owner defined, the value ZettelID(0) is returned.
func Owner() domain.ZettelID { return startupConfig.owner }

// IsOwner returns true, if the given user is the owner of the Zettelstore.
func IsOwner(zid domain.ZettelID) bool { return zid.IsValid() && zid == startupConfig.owner }

// WithAuth returns true if user authentication is enabled.
func WithAuth() bool { return startupConfig.withAuth }

// Secret returns the interal application secret. It is typically used to
// encrypt session values.
func Secret() []byte { return startupConfig.secret }

// TokenLifetime return the token lifetime for the web/HTML access and for the
// API access. If lifetime for API access is equal to zero, no API access is
// possible.
func TokenLifetime() (htmlLifetime, apiLifetime time.Duration) {
	return startupConfig.htmlLifetime, startupConfig.apiLifetime
}
