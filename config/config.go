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
	"strconv"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/place/stock"
)

// --- Configuration zettel --------------------------------------------------

var configStock stock.Stock

// SetupConfiguration enables the configuration data.
func SetupConfiguration(place place.Place) {
	if configStock != nil {
		panic("configStock already set")
	}
	configStock = stock.NewStock(place)
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

// GetExpertMode returns the current value of the "expert-mode" key
func GetExpertMode() bool {
	if config := getConfigurationMeta(); config != nil {
		if mode, ok := config.Get(domain.MetaKeyExpertMode); ok {
			return domain.BoolValue(mode)
		}
	}
	return false
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

// GetStart returns the value of the "start" key.
func GetStart() domain.ZettelID {
	if config := getConfigurationMeta(); config != nil {
		if start, ok := config.Get(domain.MetaKeyStart); ok {
			if startID, err := domain.ParseZettelID(start); err == nil {
				return startID
			}
		}
	}
	return domain.InvalidZettelID
}

// GetDefaultVisibility returns the default value for zettel visibility.
func GetDefaultVisibility() domain.Visibility {
	if config := getConfigurationMeta(); config != nil {
		if value, ok := config.Get(domain.MetaKeyDefaultVisibility); ok {
			if vis := domain.GetVisibility(value); vis != domain.VisibilityUnknown {
				return vis
			}
		}
	}
	return domain.VisibilityLogin
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

// GetMarkerExternal returns the current value of the "marker-external" key.
func GetMarkerExternal() string {
	if config := getConfigurationMeta(); config != nil {
		if html, ok := config.Get(domain.MetaKeyMarkerExternal); ok {
			return html
		}
	}
	return "&#8599;&#xfe0e;"
}

// GetFooterHTML returns HTML code tht should be embedded into the footer of each WebUI page
func GetFooterHTML() string {
	if config := getConfigurationMeta(); config != nil {
		if data, ok := config.Get(domain.MetaKeyFooterHTML); ok {
			return data
		}
	}
	return ""
}

// GetListPageSize returns the maximum length of a list to be returned in WebUI.
// A value less or equal to zero signals no limit.
func GetListPageSize() int {
	if config := getConfigurationMeta(); config != nil {
		if data, ok := config.Get(domain.MetaKeyListPageSize); ok {
			if value, err := strconv.Atoi(data); err == nil {
				return value
			}
		}
	}
	return 0
}
