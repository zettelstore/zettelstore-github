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

// Package main is the starting point for the zettel parser command.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	_ "zettelstore.de/z/cmd"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
)

func parse(name string, format string) {
	f, err := os.Open(name)
	if err != nil {
		fmt.Println("Open:", err)
		return
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("Read:", err)
		return
	}

	data := string(src)
	src = nil
	inp := input.NewInput(data)
	meta := domain.NewMetaFromInput("", inp)
	z, _ := parser.ParseZettel(
		domain.Zettel{
			Meta:    meta,
			Content: domain.NewContent(data[inp.Pos:]),
		},
		getSyntax(meta),
	)
	if enc := encoder.Create(format); enc != nil {
		if _, err := enc.WriteZettel(os.Stdout, z); err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unknown format %q\n", format)
	}
}

func getSyntax(meta *domain.Meta) string {
	if syntax, ok := meta.Get(domain.MetaKeySyntax); ok {
		return syntax
	}
	return config.Config.GetDefaultSyntax()
}

func main() {
	var format string
	flag.StringVar(&format, "t", "html", "target output format")
	flag.Parse()
	for _, n := range flag.Args() {
		parse(n, format)
	}
}
