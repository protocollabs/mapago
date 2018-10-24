package mods

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type config struct {
	Port uint16 `json:"port"`
	Addr string `json:"addr"`
}

func NewConfigClient() config {
	return config{
		Port: 6666,
		Addr: "::1",
	}
}

func (m *modUdpPingClient) configOverwriteByUserConf(user config) {
	if user.Addr != "" {
		m.config.Addr = user.Addr
	}
	if user.Port != 0 {
		m.config.Port = user.Port
	}
}

func (m *modUdpPingClient) configOverwriteByArgsConf() {
	// overwrite based on args
	for k, v := range m.cli_args {
		fmt.Fprintf(os.Stderr, "%+v     %+v\n", k, v)
		if k == "addr" {
			m.config.Addr = v
		}
		if k == "port" {
			val, err := strconv.ParseUint(v, 0, 16)
			if err != nil {
				panic("to large for port")
			}
			m.config.Port = uint16(val)
		}
	}
}

type modUdpPingClient struct {
	cli_args map[string]string
	verbose  bool
	config   config
}

// Default Values
func NewModUdpPingClient() *modUdpPingClient {
	return &modUdpPingClient{
		verbose:  false,
		cli_args: make(map[string]string),
		config:   NewConfigClient(),
	}
}

func (m *modUdpPingClient) handleConfigFile(filename string) (config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Config file is not readable %s\n", filename)
		}
		return config{}, err
	}

	var c config
	err = json.Unmarshal(content, &c)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	return c, nil
}

// first, check if a Config file is given, if so parse and overwrite defaults,
// later, check for arguments, if available and valid, overwrite defaults and
// Config. Arguments has the highest precedence
func (m *modUdpPingClient) Init() error {
	if m.verbose {
		fmt.Fprintf(os.Stdout, "Default config:\n")
		internal_conf := m.configJsonify()
		fmt.Fprintf(os.Stdout, "%+v\n", internal_conf)
	}

	// check if config file is available
	if filename, ok := m.cli_args["config"]; ok {
		config, err := m.handleConfigFile(filename)
		if err != nil {
			panic("fixme: implement me")
			return err
		}

		fmt.Fprintf(os.Stderr, "Config: %+v\n", config)
		m.configOverwriteByUserConf(config)
	}

	m.configOverwriteByArgsConf()

	// before we start we print the final config, if request
	if m.verbose {
		fmt.Fprintf(os.Stdout, "Final Config:\n")
		internal_conf := m.configJsonify()
		fmt.Fprintf(os.Stdout, "%+v\n", internal_conf)
	}

	return nil
}

func (m *modUdpPingClient) Start() error {
	return nil
}

func valid_option(s string) bool {
	return strings.Contains(s, "=")
}

func (m *modUdpPingClient) configJsonify() string {
	b, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}

// parse simple parse command line and safe for later use
// parse do not overwrite the defaults, this is done later
// in init, where sanity checks are done too
func (m *modUdpPingClient) Parse() error {

	listCommand := flag.NewFlagSet("", flag.ExitOnError)
	flagVerbose := listCommand.Bool("verbose", false, "verbose output")
	flagPrintConfig := listCommand.Bool("print-config", false, "print default config")

	argcount := 3
	listCommand.Parse(os.Args[argcount:])
	if *flagVerbose {
		m.verbose = true
		fmt.Fprintf(os.Stderr, "verbose:       enabled\n")
		argcount += 1
	}
	if *flagPrintConfig {
		config := m.configJsonify()
		fmt.Fprintf(os.Stderr, "default config:\n")
		fmt.Fprintf(os.Stdout, "%s\n", config)
		os.Exit(0)
	}

	// parse arguments
	for _, word := range os.Args[argcount:] {
		ok := valid_option(word)
		if !ok {
			return fmt.Errorf("not a valid option: %q, must be key=val", word)
		}
		kv := strings.Split(word, "=")
		m.cli_args[kv[0]] = kv[1]
	}

	return nil
}
