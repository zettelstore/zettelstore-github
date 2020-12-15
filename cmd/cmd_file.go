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
	format := cfg.GetDefault(config.StartupKeyTargetFormat, "html")
	meta, inp, err := getInput(cfg)
	if meta == nil {
		return 2, err
	}
	z := parser.ParseZettel(
		domain.Zettel{
			Meta:    meta,
			Content: domain.NewContent(inp.Src[inp.Pos:]),
		},
		config.GetSyntax(meta),
	)
	enc := encoder.Create(
		format,
		&encoder.StringOption{Key: "lang", Value: config.GetLang(meta)},
	)
	if enc == nil {
		fmt.Fprintf(os.Stderr, "Unknown format %q\n", format)
		return 2, nil
	}
	_, err = enc.WriteZettel(os.Stdout, z, format != "raw")
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
