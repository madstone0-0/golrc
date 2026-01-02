package cli

import (
	"flag"
	"golrc/internal/appstate"
	"golrc/internal/logger"
)

var (
	log = logger.NewTaggedLogger("CLI")
)

type Args struct {
	Version        bool   // Display version information
	Directory      string // Music directory
	FilterExisting bool   // Filter out music files that already have lyrics
	DryRun         bool   // Perform a trial run with no changes made

	Debug  bool // Run in debug mode with logs
	NArgs  int  // Number of arguments passed
	NFlags int  // Number of flags passed
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
		version        bool
		directory      string
		filterExisting bool
		dryRun         bool

		debug bool
	)

	nonDefaultFlags := []string{"dir"}

	flag.BoolVar(&version, "version", false, "Display version information")
	flag.BoolVar(&version, "v", false, "Display version information")
	flag.StringVar(&directory, "dir", "", "Directory to search for music")
	flag.BoolVar(&filterExisting, "exist", false, "Filter out music files that already have lyrics")
	flag.BoolVar(&filterExisting, "e", false, "Filter out music files that already have lyrics")
	flag.BoolVar(&dryRun, "dry", false, "Perform a trial run with no changes made")
	flag.BoolVar(&dryRun, "n", false, "Perform a trial run with no changes made")

	flag.BoolVar(&debug, "debug", false, "Run in debug mode with logs")
	flag.BoolVar(&debug, "d", false, "Run in debug mode with logs")

	flag.Parse()
	if !version && (flag.NFlag() > 0 || flag.NArg() > 0) {
		for _, f := range nonDefaultFlags {
			if !isFlagPassed(f) {
				log.F("Non-default flag not passed", "flag", f)
			}
		}
	}

	logger.DEBUG = debug
	appstate.DRY_RUN = dryRun

	log.D("Parsed command line args", "args", map[string]any{
		"version":        version,
		"debug":          debug,
		"directory":      directory,
		"dryRun":         dryRun,
		"filterExisting": filterExisting,
		"nArgs":          flag.NArg(),
		"nFlags":         flag.NFlag(),
	},
	)

	return Args{
		Version:        version,
		Directory:      directory,
		FilterExisting: filterExisting,
		DryRun:         dryRun,

		NArgs:  flag.NArg(),
		NFlags: flag.NFlag(),
		Debug:  debug,
	}, nil
}

func PrintUsage() {
	flag.Usage()
}
