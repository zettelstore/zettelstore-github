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
	"os"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"

	"zettelstore.de/z/domain"
)

// ---------- Subcommand: password -------------------------------------------

func cmdPassword(cfg *domain.Meta) (int, error) {
	password, err := getPassword("Password")
	if err != nil {
		return 2, err
	}
	passwordAgain, err := getPassword("   Again")
	if err != nil {
		return 2, err
	}
	if string(password) != string(passwordAgain) {
		fmt.Fprintln(os.Stderr, "Passwords differ!")
		return 2, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return 2, err
	}
	fmt.Printf("%s\n", hashedPassword)
	return 0, nil
}

func getPassword(prompt string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	return password, err
}
