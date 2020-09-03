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
	"io/ioutil"
	"os"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/input"
	"zettelstore.de/z/parser"
)

// ---------- Subcommand: file -----------------------------------------------

func cmdFile(cfg *domain.Meta) (int, error) {
	format := cfg.GetDefault("target-format", "html")
	meta, inp, err := getInput(cfg)
	if meta == nil {
		return 2, err
	}
	z, _ := parser.ParseZettel(
		domain.Zettel{
			Meta:    meta,
			Content: domain.NewContent(inp.Src[inp.Pos:]),
		},
		config.Config.GetSyntax(meta),
	)
	enc := encoder.Create(
		format,
		&encoder.StringOption{Key: "lang", Value: config.Config.GetLang(meta)},
	)
	if enc == nil {
		fmt.Fprintf(os.Stderr, "Unknown format %q\n", format)
		return 2, nil
	}
	_, err = enc.WriteZettel(os.Stdout, z)
	if err != nil {
		return 2, err
	}
	fmt.Println()

	return 0, nil
}

func getInput(cfg *domain.Meta) (*domain.Meta, *input.Input, error) {
	name, ok := cfg.Get("arg-1")
	if !ok {
		src, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, nil, err
		}
		inp := input.NewInput(string(src))
		meta := domain.NewMetaFromInput(domain.NewZettelID(true), inp)
		return meta, inp, nil
	}

	src, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, nil, err
	}
	inp := input.NewInput(string(src))
	meta := domain.NewMetaFromInput(domain.NewZettelID(true), inp)

	name2, ok := cfg.Get("arg-2")
	if ok {
		src, err := ioutil.ReadFile(name2)
		if err != nil {
			return nil, nil, err
		}
		inp = input.NewInput(string(src))
	}
	return meta, inp, nil
}
