package main

import (
	"fmt"
	"golrc/cli"
	"golrc/internal/appstate/constant"
	log "golrc/internal/logger"
	"golrc/internal/utils"
	"golrc/library"
	"golrc/provider"
	"golrc/provider/lrclib"
	"os"
)

func main() {
	args, err := cli.GetArgs()
	if err != nil {
		log.F("Error parsing arguments", "error", err)
	}
	if args.NArgs == 0 && args.NFlags == 0 {
		cli.PrintUsage()
		os.Exit(1)
	}

	if args.Version {
		fmt.Printf("%s version: %s\n", constant.APPNAME, constant.VERSION)
		os.Exit(0)
	}

	if args.DryRun {
		log.I("Dry run mode enabled. No files will be written.")
	}

	musicDir, err := utils.ResolvePath(args.Directory)
	if err != nil {
		log.F("Error resolving music directory", "error", err)
	}

	files, err := library.ReadMusicDir(musicDir, args.FilterExisting)
	if err != nil {
		log.F("Failed to read music directory", "error", err)
	}
	log.I("[+] Found music files", "count", len(files))

	tags, err := library.ParseTags(files)
	if err != nil {
		log.F("Failed to parse tags", "error", err)
	}
	log.I("[+] Parsed tags", "count", len(tags))

	switch args.Provider {
	case cli.AUTO:
		ctors := []provider.ProviderCtor{
			{
				Name:     "LrcLib",
				Priority: 0,
				Ctor: func() (provider.Provider, error) {
					return lrclib.New()
				},
			},
		}

		log.I("[+] Using provider", "provider", args.Provider.String())
		err := provider.TryProviders(tags, ctors...)
		if err != nil {
			log.F("Error trying providers", "error", err)
		}

	case cli.LRCLIB:
		prv, err := lrclib.New()
		if err != nil {
			log.F("Failed to create LrcLib provider", "error", err)
		}
		log.I("[+] Using provider", "provider", args.Provider)
		err = provider.TryProvider(tags, prv)
		if err != nil {
			log.F("Error trying LrcLib provider", "error", err)
		}
	default:
		log.F("Unsupported lyrics provider", "provider", args.Provider)
	}

}
