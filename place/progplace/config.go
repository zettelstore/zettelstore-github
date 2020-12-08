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
	"strings"

	"zettelstore.de/z/domain"
)

func genConfigM(zid domain.ZettelID) *domain.Meta {
	if myPlace.startConfig == nil {
		return nil
	}
	meta := domain.NewMeta(zid)
	meta.Set(domain.MetaKeyTitle, "Zettelstore Startup Configuration")
	meta.Set(domain.MetaKeyRole, "configuration")
	meta.Set(domain.MetaKeySyntax, "zmk")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityOwner)
	meta.Set(domain.MetaKeyReadOnly, "true")
	return meta
}

func genConfigC(meta *domain.Meta) string {
	var sb strings.Builder
	for i, p := range myPlace.startConfig.Pairs() {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("; ''")
		sb.WriteString(p.Key)
		sb.WriteString("''")
		if p.Value != "" {
			sb.WriteString("\n: ``")
			for _, r := range p.Value {
				if r == '`' {
					sb.WriteByte('\\')
				}
				sb.WriteRune(r)
			}
			sb.WriteString("``")
		}
	}
	return sb.String()
}
