package cli

import (
	"flag"
	"golrc/internal/logger"
)

var (
	log = logger.NewTaggedLogger("CLI")
)

type Args struct {
	Version   bool   // Display version information
	Directory string // Music directory

	Debug bool // Run in debug mode with logs
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

func GetArgs() (Args, error) {
	var (
		version   bool
		directory string

		debug bool
	)

	nonDefaultFlags := []string{"directory"}

	flag.BoolVar(&version, "version", false, "Display version information")
	flag.BoolVar(&version, "v", false, "Display version information")
	flag.StringVar(&directory, "dir", "", "Directory to search for music")

	flag.BoolVar(&debug, "debug", false, "Run in debug mode with logs")
	flag.BoolVar(&debug, "d", false, "Run in debug mode with logs")

	flag.Parse()
	for _, f := range nonDefaultFlags {
		if !isFlagPassed(f) {
			log.F("Non-default flag not passed", "flag", f)
		}
	}

	logger.DEBUG = debug

	log.D("Parsed command line args", "args", map[string]any{
		"version":   version,
		"debug":     debug,
		"directory": directory,
	},
	)

	return Args{
		Version:   version,
		Directory: directory,

		Debug: debug,
	}, nil
}
