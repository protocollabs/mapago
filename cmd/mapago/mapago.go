package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/protocollabs/mapago/core"
	"github.com/protocollabs/mapago/core/log"
	"github.com/protocollabs/mapago/mods"
)

type Mod interface {
	Parse() error
	Init() error
	Start() error
}

// campaign has to provide the same functions as
// Modules. Simple typedef then
type Campaign Mod

var (
	errProjectNotFound    = fmt.Errorf("could not find project")
)

func parse_args() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}

	return "", fmt.Errorf("list or count subcommand is required")
}

func usage() {
	fmt.Fprintf(os.Stderr, "invalid arg, must be mod- or campaign-\n")
}

func is_valid_mode(s string) bool {
	if s == "server" || s == "client" {
		return true
	}
	return false
}

func mode_mod(name string) {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "moaare args required")
		os.Exit(0)
	}

	// must be server or client
	mode := os.Args[2]
	if !is_valid_mode(mode) {
		fmt.Fprintf(os.Stderr, "not a valid mode: %s (server or client)", mode)
		os.Exit(0)
	}

	var mod Mod
	if name == "mod-udp-ping" {
		if mode == "client" {
			mod = mods.NewModUdpPingClient()
		} else {
			panic("fixme: implement me")
		}
	} else {
		panic("fixme: implement me")
	}

	err := mod.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	err = mod.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func main() {
	var log = log.New()
	log.Debugf("xx")
	fmt.Fprintf(os.Stderr, "mapago(c) - 2017,2018\n")
	fmt.Fprintf(os.Stderr, "build version: %s\n", core.BuildVersion)
	fmt.Fprintf(os.Stderr, "build date:    %s\n", core.BuildDate)

	name, err := parse_args()
	if err != nil {
		usage()
		os.Exit(1)
	}

	if strings.HasPrefix(name, "mod-") {
		mode_mod(name)
	} else if strings.HasPrefix(name, "campaign-") {
		//mode_campaign(name)
	} else {
		usage()
		os.Exit(1)
	}

	os.Exit(0)
}
