package main

import (
	"fmt"
	"golrc/cli"
	"golrc/internal/appstate"
	log "golrc/internal/logger"
	"os"
)

func main() {
	args, err := cli.GetArgs()
	if err != nil {
		log.F("Error parsing arguments", "error", err)
	}

	if args.Version {
		fmt.Printf("%s version: %s\n", appstate.APPNAME, appstate.VERSION)
		os.Exit(0)
	}

	fmt.Println(appstate.APPNAME)
}
