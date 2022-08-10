package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
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
	var path_config string

	if len(os.Args) <= 1 {
		fmt.Printf("usage: %s [subcommand]\n", filepath.Base(os.Args[0]))
		fmt.Println("subcommand is either `measure', `serve', or `help'`.")
		os.Exit(1)
	}

	cmd_measure := flag.NewFlagSet("measure", flag.ExitOnError)
	cmd_measure.StringVar(&path_config, FLAG_CONFIG_PATH, DEFAULT_CONFIG_PATH, HELP_CONFIG_PATH)

	cmd_serve := flag.NewFlagSet("serve", flag.ExitOnError)
	cmd_serve.StringVar(&path_config, FLAG_CONFIG_PATH, DEFAULT_CONFIG_PATH, HELP_CONFIG_PATH)

	switch os.Args[1] {
	case "measure":
		cmd_measure.Parse(os.Args[2:])
		make_sure_not_root()
		measure(path_config)
	case "serve":
		cmd_serve.Parse(os.Args[2:])
		make_sure_not_root()
		serve(path_config)
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
