package cli

import (
	"errors"
	"flag"
	"golrc/internal/appstate"
	"golrc/internal/logger"
	"strings"
)

var (
	log = logger.NewTaggedLogger("CLI")
)

type ProviderType int

const (
	AUTO ProviderType = iota
	LRCLIB
)

func (p ProviderType) String() string {
	switch p {
	case LRCLIB:
		return "lrclib"
	case AUTO:
		return "auto"
	default:
		return ""
	}
}

func parseProviderType(provider string) (ProviderType, error) {
	switch strings.ToLower(provider) {
	case "lrclib":
		return LRCLIB, nil
	case "auto":
		return AUTO, nil
	default:
		return -1, errors.New("invalid provider type")
	}
}

type Args struct {
	Version        bool         // Display version information
	Directory      string       // Music directory
	FilterExisting bool         // Filter out music files that already have lyrics
	DryRun         bool         // Perform a trial run with no changes made
	Provider       ProviderType // Lyrics provider to use
	MaxConc        int          // Maximum number of lyrics to fetch at once

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

type Validator = func() error

func V[T any](arg T, val func(T) error) Validator {
	return func() error {
		return val(arg)
	}
}

func Validate(vs ...Validator) error {
	var err error

	for _, v := range vs {
		e := v()
		if e != nil {
			err = errors.Join(e, err)
		}
	}

	return err
}

func GetArgs() (Args, error) {
	var (
		version        bool
		directory      string
		filterExisting bool
		dryRun         bool
		maxConc        int

		provider ProviderType

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
	flag.IntVar(&maxConc, "mC", 5, "Fetch n lyrics concurrently")
	flag.Func("provider", "Lyrics provider to use (lrclib)", func(s string) error {
		p, err := parseProviderType(s)
		if err != nil {
			return err
		}
		provider = p
		return nil
	})

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

	err := Validate(V(maxConc, func(maxConc int) error {
		if maxConc <= 0 || maxConc >= 30 {
			return errors.New("max concurrent fetches must be between 1 and 30")
		}
		return nil
	}))

	if err != nil {
		return Args{}, err
	}

	logger.DEBUG = debug
	appstate.DRY_RUN = dryRun

	log.D("Parsed command line args", "args", map[string]any{
		"version":        version,
		"debug":          debug,
		"directory":      directory,
		"dryRun":         dryRun,
		"maxConc":        maxConc,
		"filterExisting": filterExisting,
		"provider":       provider,
		"nArgs":          flag.NArg(),
		"nFlags":         flag.NFlag(),
	},
	)

	return Args{
		Version:        version,
		Directory:      directory,
		FilterExisting: filterExisting,
		DryRun:         dryRun,
		MaxConc:        maxConc,
		Provider:       provider,

		NArgs:  flag.NArg(),
		NFlags: flag.NFlag(),
		Debug:  debug,
	}, nil
}

func PrintUsage() {
	flag.Usage()
}
