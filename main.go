package main

import (
	"fmt"
	"golrc/cli"
	"golrc/internal/appstate/constant"
	log "golrc/internal/logger"
	"golrc/internal/utils"
	"golrc/library"
	"golrc/provider/lrclib"
	"os"
)

func main() {
	args, err := cli.GetArgs()
	if err != nil {
		log.F("Error parsing arguments", "error", err)
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

	tracks, err := lrclib.GetAllLyrics(tags)
	if err != nil {
		log.F("Failed to get lyrics", "error", err)
	}
	log.I("[+] Fetched lyrics", "count", len(tracks))
	log.D("[+] Tracks with lyrics", "tracks", tracks)

	wrote := lrclib.WriteLyrics(tracks)

	log.I("[+] Lyrics fetching completed", "total", len(files), "written", wrote)

}
