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

// Package config provides functions to retrieve configuration data.
package config

import (
	"fmt"
	"hash/fnv"
	"os"
	"runtime"
	"strconv"
	"time"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/stock"
)

// Version describes all elements of a software version.
type Version struct {
	Prog      string // Name of the software
	Build     string // Representation of build process
	Hostname  string // Host name a reported by the kernel
	GoVersion string // Version of go
	Os        string // GOOS
	Arch      string // GOARCH
	// More to come
}

var version Version

// SetupVersion initializes the version data.
func SetupVersion(progName, buildVersion string) {
	version.Prog = progName
	if buildVersion == "" {
		version.Build = "unknown"
	} else {
		version.Build = buildVersion
	}
	if hn, err := os.Hostname(); err == nil {
		version.Hostname = hn
	} else {
		version.Hostname = "*unknown host*"
	}
	version.GoVersion = runtime.Version()
	version.Os = runtime.GOOS
	version.Arch = runtime.GOARCH
}

// GetVersion returns the current software version data.
func GetVersion() Version { return version }

// --- Startup config --------------------------------------------------------

var startupConfig struct {
	readonly     bool
	urlPrefix    string
	insecCookie  bool
	owner        domain.ZettelID
	withAuth     bool
	secret       []byte
	htmlLifetime time.Duration
	apiLifetime  time.Duration
}

// SetupStartup initializes the startup data.
func SetupStartup(cfg *domain.Meta) {
	if startupConfig.urlPrefix != "" {
		panic("startupConfig already set")
	}
	startupConfig.readonly = cfg.GetBool("readonly")
	startupConfig.urlPrefix = cfg.GetDefault("url-prefix", "/")
	startupConfig.owner = domain.InvalidZettelID
	if owner, ok := cfg.Get("owner"); ok {
		if zid, err := domain.ParseZettelID(owner); err == nil {
			startupConfig.owner = zid
			startupConfig.withAuth = true
		}
	}
	if startupConfig.withAuth {
		startupConfig.insecCookie = cfg.GetBool("insecure-cookie")
		startupConfig.secret = calcSecret(cfg)
		startupConfig.htmlLifetime = getDuration(
			cfg, "token-lifetime-html", 1*time.Hour, 1*time.Minute, 30*24*time.Hour)
		startupConfig.apiLifetime = getDuration(
			cfg, "token-lifetime-api", 10*time.Minute, 0, 1*time.Hour)
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

// IsReadOnly returns whether the system is in read-only mode or not.
func IsReadOnly() bool { return startupConfig.readonly }

// URLPrefix returns the configured prefix to be used when providing URL to
// the service.
func URLPrefix() string { return startupConfig.urlPrefix }

// SecureCookie returns whether the web app should set cookies to secure mode.
func SecureCookie() bool { return !startupConfig.insecCookie }

// Owner returns the zid of the zettelkasten's owner.
// If there is no owner defined, the value ZettelID(0) is returned.
func Owner() domain.ZettelID { return startupConfig.owner }

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

// --- Configuration zettel --------------------------------------------------

var configStock stock.Stock

// SetupConfiguration enables the configuration data.
func SetupConfiguration(store store.Store) {
	if configStock != nil {
		panic("configStock already set")
	}
	configStock = stock.NewStock(store)
	if err := configStock.Subscribe(domain.ConfigurationID); err != nil {
		panic(err)
	}
}

// getConfigurationMeta returns the meta data of the configuration zettel.
func getConfigurationMeta() *domain.Meta {
	if configStock == nil {
		panic("configStock not set")
	}
	return configStock.GetMeta(domain.ConfigurationID)
}

// GetDefaultTitle returns the current value of the "default-title" key.
func GetDefaultTitle() string {
	if config := getConfigurationMeta(); config != nil {
		if title, ok := config.Get(domain.MetaKeyDefaultTitle); ok {
			return title
		}
	}
	return "Untitled"
}

// GetDefaultSyntax returns the current value of the "default-syntax" key.
func GetDefaultSyntax() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			if syntax, ok := config.Get(domain.MetaKeyDefaultSyntax); ok {
				return syntax
			}
		}
	}
	return "zmk"
}

// GetDefaultRole returns the current value of the "default-role" key.
func GetDefaultRole() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			if role, ok := config.Get(domain.MetaKeyDefaultRole); ok {
				return role
			}
		}
	}
	return "zettel"
}

// GetDefaultLang returns the current value of the "default-lang" key.
func GetDefaultLang() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			if lang, ok := config.Get(domain.MetaKeyDefaultLang); ok {
				return lang
			}
		}
	}
	return "en"
}

// GetDefaultCopyright returns the current value of the "default-copyright" key.
func GetDefaultCopyright() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			if copyright, ok := config.Get(domain.MetaKeyDefaultCopyright); ok {
				return copyright
			}
		}
		// TODO: get owner
	}
	return ""
}

// GetDefaultLicense returns the current value of the "default-license" key.
func GetDefaultLicense() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			if license, ok := config.Get(domain.MetaKeyDefaultLicense); ok {
				return license
			}
		}
	}
	return ""
}

// GetSiteName returns the current value of the "site-name" key.
func GetSiteName() string {
	if config := getConfigurationMeta(); config != nil {
		if name, ok := config.Get(domain.MetaKeySiteName); ok {
			return name
		}
	}
	return "Zettelstore"
}

// GetYAMLHeader returns the current value of the "yaml-header" key.
func GetYAMLHeader() bool {
	if config := getConfigurationMeta(); config != nil {
		return config.GetBool(domain.MetaKeyYAMLHeader)
	}
	return false
}

// GetZettelFileSyntax returns the current value of the "zettel-file-syntax" key.
func GetZettelFileSyntax() []string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetListOrNil(domain.MetaKeyZettelFileSyntax)
	}
	return nil
}

// GetIconMaterial returns the current value of the "icon-material" key.
func GetIconMaterial() string {
	if config := getConfigurationMeta(); config != nil {
		if html, ok := config.Get(domain.MetaKeyIconMaterial); ok {
			return html
		}
	}
	return fmt.Sprintf(
		"<img class=\"zs-text-icon\" src=\"%vc/%v\">",
		URLPrefix(),
		domain.MaterialIconID.Format())
}

var mapDefaultKeys = map[string]func() string{
	domain.MetaKeyCopyright: GetDefaultCopyright,
	domain.MetaKeyLang:      GetDefaultLang,
	domain.MetaKeyLicense:   GetDefaultLicense,
	domain.MetaKeyRole:      GetDefaultRole,
	domain.MetaKeySyntax:    GetDefaultSyntax,
	domain.MetaKeyTitle:     GetDefaultTitle,
}

// AddDefaultValues enriches the given meta data with its default values.
func AddDefaultValues(meta *domain.Meta) *domain.Meta {
	result := meta
	for k, f := range mapDefaultKeys {
		if _, ok := result.Get(k); !ok {
			if result == meta {
				result = meta.Clone()
			}
			if val := f(); len(val) > 0 || meta.Type(k) == domain.MetaTypeEmpty {
				result.Set(k, val)
			}
		}
	}
	if result != meta && meta.IsFrozen() {
		result.Freeze()
	}
	return result
}

// GetSyntax returns the value of the "syntax" key of the given meta. If there
// is no such value, GetDefaultLang is returned.
func GetSyntax(meta *domain.Meta) string {
	if syntax, ok := meta.Get(domain.MetaKeySyntax); ok && len(syntax) > 0 {
		return syntax
	}
	return GetDefaultSyntax()
}

// GetLang returns the value of the "lang" key of the given meta. If there is
// no such value, GetDefaultLang is returned.
func GetLang(meta *domain.Meta) string {
	if lang, ok := meta.Get(domain.MetaKeyLang); ok && len(lang) > 0 {
		return lang
	}
	return GetDefaultLang()
}
