//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package progplace provides zettel that inform the user about the internal Zettelstore state.
package progplace

import (
	"fmt"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

func getVersionMeta(zid domain.ZettelID, title string) *domain.Meta {
	meta := domain.NewMeta(zid)
	meta.Set(domain.MetaKeyTitle, title)
	meta.Set(domain.MetaKeyRole, "configuration")
	meta.Set(domain.MetaKeySyntax, "zmk")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityLogin)
	meta.Set(domain.MetaKeyReadOnly, "true")
	return meta
}

func genVersionBuildM(zid domain.ZettelID) *domain.Meta {
	meta := getVersionMeta(zid, "Zettelstore Version")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityPublic)
	return meta
}
func genVersionBuildC(meta *domain.Meta) string { return config.GetVersion().Build }

func genVersionHostM(zid domain.ZettelID) *domain.Meta {
	return getVersionMeta(zid, "Zettelstore Host")
}
func genVersionHostC(meta *domain.Meta) string { return config.GetVersion().Hostname }

func genVersionOSM(zid domain.ZettelID) *domain.Meta {
	return getVersionMeta(zid, "Zettelstore Operating System")
}
func genVersionOSC(meta *domain.Meta) string {
	v := config.GetVersion()
	return fmt.Sprintf("%v/%v", v.Os, v.Arch)
}

func genVersionGoM(zid domain.ZettelID) *domain.Meta {
	return getVersionMeta(zid, "Zettelstore Go Version")
}
func genVersionGoC(meta *domain.Meta) string { return config.GetVersion().GoVersion }
