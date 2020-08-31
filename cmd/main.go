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
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
)

const (
	defConfigfile = ".zscfg"
)

func init() {
	RegisterCommand(Command{
		Name: "help",
		Func: func(cfg *domain.Meta) (int, error) {
			fmt.Println("Available commands:")
			for _, name := range List() {
				fmt.Printf("- %q\n", name)
			}
			return 0, nil
		},
	})
	RegisterCommand(Command{
		Name: "version",
		Func: func(cfg *domain.Meta) (int, error) {
			version := config.Config.GetVersion()
			fmt.Printf("%v, build %v\n", version.Prog, version.Build)
			return 0, nil
		},
	})
	RegisterCommand(Command{
		Name: "run",
		Func: runFunc,
		Flags: func(fs *flag.FlagSet) {
			fs.String("c", defConfigfile, "configuration file")
			fs.Uint("p", 23123, "port number")
			fs.String("d", "", "zettel directory")
			fs.Bool("r", false, "system-wide read-only mode")
			fs.Bool("v", false, "verbose mode")
		},
	})
	RegisterCommand(Command{
		Name: "file",
		Func: cmdFile,
		Flags: func(fs *flag.FlagSet) {
			fs.String("t", "html", "target output format")
		},
	})
}

func getConfig(fs *flag.FlagSet) (cfg *domain.Meta) {
	var configFile string
	if configFlag := fs.Lookup("c"); configFlag != nil {
		configFile = configFlag.Value.String()
	} else {
		configFile = defConfigfile
	}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		cfg = domain.NewMeta(domain.InvalidZettelID)
	} else {
		cfg = domain.NewMetaFromInput(domain.InvalidZettelID, input.NewInput(string(content)))
	}
	fs.Visit(func(flg *flag.Flag) {
		switch flg.Name {
		case "p":
			cfg.Set("listen-addr", "127.0.0.1:"+flg.Value.String())
		case "d":
			cfg.Set("store-1-dir", flg.Value.String())
		case "r":
			cfg.Set("readonly", flg.Value.String())
		case "v":
			cfg.Set("verbose", flg.Value.String())
		case "t":
			cfg.Set("target-format", flg.Value.String())
		}
	})

	if _, ok := cfg.Get("listen-addr"); !ok {
		cfg.Set("listen-addr", "127.0.0.1:23123")
	}
	if _, ok := cfg.Get("readonly"); !ok {
		cfg.Set("readonly", "false")
	}
	if _, ok := cfg.Get("verbose"); !ok {
		cfg.Set("verbose", "false")
	}
	if prefix, ok := cfg.Get("url-prefix"); !ok || len(prefix) == 0 || prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		cfg.Set("url-prefix", "/")
	}

	for i, arg := range fs.Args() {
		cfg.Set(fmt.Sprintf("arg-%d", i+1), arg)
	}
	return cfg
}

func executeCommand(name string, args ...string) {
	command, ok := Get(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", name)
		os.Exit(1)
	}
	fs := command.GetFlags()
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to parse flags: %v %v\n", name, args, err)
		os.Exit(1)
	}
	cfg := getConfig(fs)
	cfg.Set("command-name", name)
	config.SetupStartup(cfg)
	exitCode, err := command.Func(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
	}
	os.Exit(exitCode)
}

// Main is the real entrypoint of the zettelstore.
func Main(progName, buildVersion string) {
	config.SetupVersion(progName, buildVersion)
	if len(os.Args) <= 1 {
		executeCommand("run", "-d", "./zettel")
	} else {
		executeCommand(os.Args[1], os.Args[2:]...)
	}
}
