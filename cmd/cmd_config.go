//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package cmd

import (
	"fmt"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

// ---------- Subcommand: config ---------------------------------------------

func cmdConfig(cfg *domain.Meta) (int, error) {
	fmtVersion()
	fmt.Println("Stores")
	fmt.Printf("  Read only         = %v\n", config.IsReadOnly())
	fmt.Println("Web")
	fmt.Printf("  Listen Addr       = %q\n", cfg.GetDefault("listen-addr", "???"))
	fmt.Printf("  URL prefix        = %q\n", config.URLPrefix())
	if config.WithAuth() {
		fmt.Println("Auth")
		fmt.Printf("  Owner             = %v\n", config.Owner().Format())
		fmt.Printf("  Secure cookie     = %v\n", config.SecureCookie())
		fmt.Printf("  Persistent cookie = %v\n", config.PersistentCookie())
		htmlLifetime, apiLifetime := config.TokenLifetime()
		fmt.Printf("  HTML lifetime     = %v\n", htmlLifetime)
		fmt.Printf("  API lifetime      = %v\n", apiLifetime)
	}

	return 0, nil
}
