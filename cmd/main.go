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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
			fmtVersion()
			return 0, nil
		},
	})
	RegisterCommand(Command{
		Name:  "run",
		Func:  runFunc,
		Flags: flgRun,
	})
	RegisterCommand(Command{
		Name:  "config",
		Func:  cmdConfig,
		Flags: flgRun,
	})
	RegisterCommand(Command{
		Name: "file",
		Func: cmdFile,
		Flags: func(fs *flag.FlagSet) {
			fs.String("t", "html", "target output format")
		},
	})
	RegisterCommand(Command{
		Name: "password",
		Func: cmdPassword,
	})
}

func fmtVersion() {
	version := config.GetVersion()
	fmt.Printf("%v (%v/%v) running on %v (%v/%v)\n",
		version.Prog, version.Build, version.GoVersion,
		version.Hostname, version.Os, version.Arch)
}

func flgRun(fs *flag.FlagSet) {
	fs.String("c", defConfigfile, "configuration file")
	fs.Uint("p", 23123, "port number")
	fs.String("d", "", "zettel directory")
	fs.Bool("r", false, "system-wide read-only mode")
	fs.Bool("v", false, "verbose mode")
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
			cfg.Set(config.StartupKeyListenAddress, "127.0.0.1:"+flg.Value.String())
		case "d":
			val := flg.Value.String()
			if strings.HasPrefix(val, "/") {
				val = "dir://" + val
			} else {
				val = "dir:" + val
			}
			cfg.Set(config.StartupKeyPlaceOneURI, val)
		case "r":
			cfg.Set(config.StartupKeyReadOnlyMode, flg.Value.String())
		case "v":
			cfg.Set(config.StartupKeyVerbose, flg.Value.String())
		case "t":
			cfg.Set(config.StartupKeyTargetFormat, flg.Value.String())
		}
	})

	if _, ok := cfg.Get(config.StartupKeyListenAddress); !ok {
		cfg.Set(config.StartupKeyListenAddress, "127.0.0.1:23123")
	}
	if _, ok := cfg.Get(config.StartupKeyPlaceOneURI); !ok {
		cfg.Set(config.StartupKeyPlaceOneURI, "dir:./zettel")
	}
	if _, ok := cfg.Get(config.StartupKeyReadOnlyMode); !ok {
		cfg.Set(config.StartupKeyReadOnlyMode, "false")
	}
	if _, ok := cfg.Get(config.StartupKeyVerbose); !ok {
		cfg.Set(config.StartupKeyVerbose, "false")
	}
	if prefix, ok := cfg.Get(config.StartupKeyURLPrefix); !ok || len(prefix) == 0 || prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		cfg.Set(config.StartupKeyURLPrefix, "/")
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
		dir := "./zettel"
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create zettel directory %q (%s)\n", dir, err)
			return
		}
		executeCommand("run", "-d", dir)
	} else {
		executeCommand(os.Args[1], os.Args[2:]...)
	}
}
