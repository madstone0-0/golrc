package provider

import (
	"golrc/internal/appstate"
	"golrc/internal/logger"
	"golrc/internal/utils"
	"golrc/library"
	"path"
)

var (
	log = logger.NewTaggedLogger("PROVIDER")
)

type ITrack interface {
	GetTag() library.Tag
	GetAudioFile() string
	GetLyrics() string
}

func WriteLRC(p string, track ITrack, data string) error {
	if appstate.DRY_RUN {
		log.I("[+] Would write lyrics", "file", path.Base(p), "track", path.Base(track.GetAudioFile()))
		return nil
	}
	return utils.Write(p, data)
}
