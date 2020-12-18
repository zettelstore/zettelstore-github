//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package startup provides functions to retrieve startup configuration data.
package startup

import (
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"time"

	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
	"zettelstore.de/z/place"
)

var startupConfig struct {
	verbose       bool
	readonlyMode  bool
	urlPrefix     string
	listenAddress string
	owner         id.Zid
	withAuth      bool
	secret        []byte
	insecCookie   bool
	persistCookie bool
	htmlLifetime  time.Duration
	apiLifetime   time.Duration
	places        []string
	place         place.Place
}

// Predefined keys for startup zettel
const (
	StartupKeyInsecureCookie    = "insecure-cookie"
	StartupKeyListenAddress     = "listen-addr"
	StartupKeyOwner             = "owner"
	StartupKeyPersistentCookie  = "persistent-cookie"
	StartupKeyPlaceOneURI       = "place-1-uri"
	StartupKeyReadOnlyMode      = "read-only-mode"
	StartupKeyTokenLifetimeHTML = "token-lifetime-html"
	StartupKeyTokenLifetimeAPI  = "token-lifetime-api"
	StartupKeyURLPrefix         = "url-prefix"
	StartupKeyVerbose           = "verbose"
)

// SetupStartup initializes the startup data.
func SetupStartup(cfg *meta.Meta, withPlaces bool, lastPlace place.Place) error {
	if startupConfig.urlPrefix != "" {
		panic("startupConfig already set")
	}
	startupConfig.verbose = cfg.GetBool(StartupKeyVerbose)
	startupConfig.readonlyMode = cfg.GetBool(StartupKeyReadOnlyMode)
	startupConfig.urlPrefix = cfg.GetDefault(StartupKeyURLPrefix, "/")
	if prefix, ok := cfg.Get(StartupKeyURLPrefix); ok &&
		len(prefix) > 0 && prefix[0] == '/' && prefix[len(prefix)-1] == '/' {
		startupConfig.urlPrefix = prefix
	} else {
		startupConfig.urlPrefix = "/"
	}
	if val, ok := cfg.Get(StartupKeyListenAddress); ok {
		startupConfig.listenAddress = val // TODO: check for valid string
	} else {
		startupConfig.listenAddress = "127.0.0.1:23123"
	}
	startupConfig.owner = id.Invalid
	if owner, ok := cfg.Get(StartupKeyOwner); ok {
		if zid, err := id.Parse(owner); err == nil {
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
	if !withPlaces {
		return nil
	}
	startupConfig.places = getPlaces(cfg)
	place, err := connectPlaces(startupConfig.places, lastPlace)
	if err == nil {
		startupConfig.place = place
	}
	return err
}

func calcSecret(cfg *meta.Meta) []byte {
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

func getPlaces(cfg *meta.Meta) []string {
	hasConst := false
	var result []string = nil
	for cnt := 1; ; cnt++ {
		key := fmt.Sprintf("place-%v-uri", cnt)
		uri, ok := cfg.Get(key)
		if !ok || uri == "" {
			if cnt > 1 {
				break
			}
			uri = "dir:./zettel"
		}
		if uri == "const:" {
			hasConst = true
		}
		if IsReadOnlyMode() {
			if u, err := url.Parse(uri); err == nil {
				// TODO: the following is wrong under some circumstances:
				// 1. query parameter "readonly" is already set
				// 2. fragment is set
				if len(u.Query()) == 0 {
					uri += "?readonly"
				} else {
					uri += "&readonly"
				}
			}
		}
		result = append(result, uri)
	}
	if !hasConst {
		result = append(result, "const:")
	}
	return result
}

func connectPlaces(placeURIs []string, lastPlace place.Place) (place.Place, error) {
	if len(placeURIs) == 0 {
		return lastPlace, nil
	}
	next, err := connectPlaces(placeURIs[1:], lastPlace)
	if err != nil {
		return nil, err
	}
	p, err := place.Connect(placeURIs[0], next)
	return p, err
}

func getDuration(
	cfg *meta.Meta, key string, defDur, minDur, maxDur time.Duration) time.Duration {
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

// IsVerbose returns whether the system should be more chatty about its operations.
func IsVerbose() bool { return startupConfig.verbose }

// IsReadOnlyMode returns whether the system is in read-only mode or not.
func IsReadOnlyMode() bool { return startupConfig.readonlyMode }

// URLPrefix returns the configured prefix to be used when providing URL to
// the service.
func URLPrefix() string { return startupConfig.urlPrefix }

// ListenAddress returns the string that specifies the the network card and the ip port
// where the server listens for requests
func ListenAddress() string { return startupConfig.listenAddress }

// SecureCookie returns whether the web app should set cookies to secure mode.
func SecureCookie() bool { return !startupConfig.insecCookie }

// PersistentCookie returns whether the web app should set persistent cookies
// (instead of temporary).
func PersistentCookie() bool { return startupConfig.persistCookie }

// Owner returns the zid of the zettelkasten's owner.
// If there is no owner defined, the value ZettelID(0) is returned.
func Owner() id.Zid { return startupConfig.owner }

// IsOwner returns true, if the given user is the owner of the Zettelstore.
func IsOwner(zid id.Zid) bool { return zid.IsValid() && zid == startupConfig.owner }

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

// Places returns a list of all place URIs
func xPlaces() []string {
	result := make([]string, len(startupConfig.places))
	copy(result, startupConfig.places)
	return result
}

// Place returns the linked list of places.
func Place() place.Place { return startupConfig.place }
