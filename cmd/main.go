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
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"zettelstore.de/z/config/runtime"
	"zettelstore.de/z/config/startup"
	"zettelstore.de/z/domain/id"
	"zettelstore.de/z/domain/meta"
	"zettelstore.de/z/input"
	"zettelstore.de/z/place/progplace"
)

const (
	defConfigfile = ".zscfg"
)

func init() {
	RegisterCommand(Command{
		Name: "help",
		Func: func(*flag.FlagSet) (int, error) {
			fmt.Println("Available commands:")
			for _, name := range List() {
				fmt.Printf("- %q\n", name)
			}
			return 0, nil
		},
	})
	RegisterCommand(Command{
		Name: "version",
		Func: func(*flag.FlagSet) (int, error) {
			fmtVersion()
			return 0, nil
		},
	})
	RegisterCommand(Command{
		Name:   "run",
		Func:   runFunc,
		Places: true,
		Flags:  flgRun,
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
	version := startup.GetVersion()
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

func getConfig(fs *flag.FlagSet) (cfg *meta.Meta) {
	var configFile string
	if configFlag := fs.Lookup("c"); configFlag != nil {
		configFile = configFlag.Value.String()
	} else {
		configFile = defConfigfile
	}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		cfg = meta.NewMeta(id.Invalid)
	} else {
		cfg = meta.NewMetaFromInput(id.Invalid, input.NewInput(string(content)))
	}
	fs.Visit(func(flg *flag.Flag) {
		switch flg.Name {
		case "p":
			cfg.Set(startup.StartupKeyListenAddress, "127.0.0.1:"+flg.Value.String())
		case "d":
			val := flg.Value.String()
			if strings.HasPrefix(val, "/") {
				val = "dir://" + val
			} else {
				val = "dir:" + val
			}
			cfg.Set(startup.StartupKeyPlaceOneURI, val)
		case "r":
			cfg.Set(startup.StartupKeyReadOnlyMode, flg.Value.String())
		case "v":
			cfg.Set(startup.StartupKeyVerbose, flg.Value.String())
		}
	})
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
	err := startup.SetupStartup(cfg, command.Places, progplace.Get())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to specified places")
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(2)
	}
	if command.Places {
		if err := startup.Place().Start(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "Unable to start zettel place")
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(2)
		}
		runtime.SetupConfiguration(startup.Place())
		progplace.Setup(cfg, startup.Place())
	}
	exitCode, err := command.Func(fs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
	}
	os.Exit(exitCode)
}

// Main is the real entrypoint of the zettelstore.
func Main(progName, buildVersion string) {
	startup.SetupVersion(progName, buildVersion)
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
