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
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/stock"
)

// Version describes all elements of a software version.
type Version struct {
	Release string // Official software release version
	Build   string // Internal representation of build process
	// More to come
}

var startupConfig *domain.Meta

// SetupStartup initializes the startup data.
func SetupStartup(cfg *domain.Meta) {
	if startupConfig != nil {
		panic("startupConfig already set")
	}
	if s := cfg.GetDefault("release-version", ""); len(s) == 0 {
		cfg.Set("release-version", "unknown")
	}
	if s := cfg.GetDefault("build-version", ""); len(s) == 0 {
		cfg.Set("build-version", "unknown")
	}
	cfg.Freeze()
	startupConfig = cfg
}

// GetVersion returns the current software version data.
func (c Type) GetVersion() Version {
	return Version{
		Release: startupConfig.GetDefault("release-version", ""),
		Build:   startupConfig.GetDefault("build-version", ""),
	}
}

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

var configStock stock.Stock

// Type is the type for the config variable.
type Type struct{}

// Config is the global configuration object.
var Config Type

// getConfigurationMeta returns the meta data of the configuration zettel.
func getConfigurationMeta() *domain.Meta {
	if configStock == nil {
		panic("configStock not set")
	}
	return configStock.GetMeta(domain.ConfigurationID)
}

// IsReadOnly returns whether the system is in read-only mode or not.
func (c Type) IsReadOnly() bool { return startupConfig.GetBool("readonly") }

// GetURLPrefix returns the configured prefix to be used when providing URL to the service.
func (c Type) GetURLPrefix() string {
	return startupConfig.GetDefault("url-prefix", "/")
}

// GetDefaultTitle returns the current value of the "default-title" key.
func (c Type) GetDefaultTitle() string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetDefault(domain.MetaKeyDefaultTitle, domain.MetaValueTitle)
	}
	return domain.MetaValueTitle
}

// GetDefaultSyntax returns the current value of the "default-syntax" key.
func (c Type) GetDefaultSyntax() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			return config.GetDefault(domain.MetaKeyDefaultSyntax, domain.MetaValueSyntax)
		}
	}
	return domain.MetaValueSyntax
}

// GetDefaultRole returns the current value of the "default-role" key.
func (c Type) GetDefaultRole() string {
	if configStock != nil {
		if config := getConfigurationMeta(); config != nil {
			return config.GetDefault(domain.MetaKeyDefaultRole, domain.MetaValueRole)
		}
	}
	return domain.MetaValueRole
}

// GetDefaultLang returns the current value of the "default-lang" key.
func (c Type) GetDefaultLang() string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetDefault(domain.MetaKeyDefaultLang, domain.MetaValueLang)
	}
	return domain.MetaValueLang
}

var defIconMaterial = "<img class=\"zs-text-icon\" src=\"/c/" + string(domain.MaterialIconID) + "\">"

// GetIconMaterial returns the current value of the "icon-material" key.
func (c Type) GetIconMaterial() string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetDefault(domain.MetaKeyIconMaterial, defIconMaterial)
	}
	return defIconMaterial
}

// GetSiteName returns the current value of the "site-name" key.
func (c Type) GetSiteName() string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetDefault(domain.MetaKeySiteName, domain.MetaValueSiteName)
	}
	return domain.MetaValueSiteName
}

// GetYAMLHeader returns the current value of the "yaml-header" key.
func (c Type) GetYAMLHeader() bool {
	if config := getConfigurationMeta(); config != nil {
		return config.GetBool(domain.MetaKeyYAMLHeader)
	}
	return false
}

// GetZettelFileSyntax returns the current value of the "zettel-file-syntax" key.
func (c Type) GetZettelFileSyntax() []string {
	if config := getConfigurationMeta(); config != nil {
		return config.GetListOrNil(domain.MetaKeyZettelFileSyntax)
	}
	return nil
}
