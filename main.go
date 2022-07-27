package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	_ "github.com/glebarez/go-sqlite"
)

func make_sure_not_root() {
	// This is fairly naive now, but better than nothing.
	if syscall.Geteuid() != 0 {
		return
	}
	if os.Getenv("LILMON_PERMIT_ROOT") == "live_dangerously" {
		fmt.Println(
			`NOTE: Instead of running lilmon as root, please consider
executing your measurement scripts selectively with sudo, doas,
or something similar. Please also note that you may want something
something else to handle TLS termination for the server.`)
	} else {
		log.Println("This program will not run as root.")
		os.Exit(20)
	}
}

func main() {
	var p_measure params_measure
	var p_serve params_serve

	if len(os.Args) <= 1 {
		fmt.Printf("usage: %s [subcommand]\n", filepath.Base(os.Args[0]))
		fmt.Println("subcommand is either `measure', `serve', or `help'`.")
		os.Exit(1)
	}

	cmd_measure := flag.NewFlagSet("measure", flag.ExitOnError)
	cmd_measure.StringVar(&p_measure.db_path, FLAG_DB_PATH, DEFAULT_DB_PATH, HELP_DB_PATH)
	cmd_measure.StringVar(&p_measure.config_path, FLAG_CONFIG_PATH, DEFAULT_CONFIG_PATH, HELP_CONFIG_PATH)
	cmd_measure.StringVar(&p_measure.shell, FLAG_SHELL, DEFAULT_SHELL, HELP_SHELL)
	cmd_measure.DurationVar(&p_measure.period, FLAG_PERIOD, DEFAULT_PERIOD, HELP_PERIOD)

	cmd_serve := flag.NewFlagSet("serve", flag.ExitOnError)
	cmd_serve.StringVar(&p_serve.db_path, FLAG_DB_PATH, DEFAULT_DB_PATH, HELP_DB_PATH)
	cmd_serve.StringVar(&p_serve.config_path, FLAG_CONFIG_PATH, DEFAULT_CONFIG_PATH, HELP_CONFIG_PATH)
	cmd_serve.StringVar(&p_serve.template_path, FLAG_TEMPLATE_PATH, DEFAULT_TEMPLATE_PATH, HELP_TEMPLATE_PATH)
	cmd_serve.StringVar(&p_serve.addr, FLAG_ADDR, DEFAULT_ADDR, HELP_ADDR)

	switch os.Args[1] {
	case "measure":
		cmd_measure.Parse(os.Args[2:])
		make_sure_not_root()
		measure(&p_measure)
	case "serve":
		cmd_serve.Parse(os.Args[2:])
		make_sure_not_root()
		serve(&p_serve)
	case "help":
		fmt.Println("The subcommands are:")
		fmt.Println()
		fmt.Println("    measure          measure metrics until interrupted")
		fmt.Println("    serve            display measurements via HTTP")
		fmt.Println("    help             show this help")
		fmt.Println()
		os.Exit(0)
	default:
		fmt.Println("unknown subcommand: ", os.Args[1])
		os.Exit(2)
	}
}
