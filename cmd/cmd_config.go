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

package cmd

import (
	"fmt"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
)

// ---------- Subcommand: config ---------------------------------------------

func cmdConfig(cfg *domain.Meta) (int, error) {

	fmt.Println("Stores")
	fmt.Printf("  Read only         = %v\n", config.IsReadOnly())
	fmt.Println("Web")
	fmt.Printf("  Listen Addr       = %q\n", cfg.GetDefault("listen-addr", "???"))
	fmt.Printf("  URL prefix        = %q\n", config.URLPrefix())
	if config.WithAuth() {
		fmt.Println("Auth")
		fmt.Printf("  Owner             = %v\n", config.Owner().Format())
		fmt.Printf("  Secure cookie     = %v\n", config.SecureCookie())
		fmt.Printf("  Persistent cookie = %v\n", config.SecureCookie())
		htmlLifetime, apiLifetime := config.TokenLifetime()
		fmt.Printf("  HTML lifetime     = %v\n", htmlLifetime)
		fmt.Printf("  API lifetime      = %v\n", apiLifetime)
	}

	return 0, nil
}
