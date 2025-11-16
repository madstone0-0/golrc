package lrclib

import (
	"errors"
	"fmt"
	"golrc/internal/fetch"
	"golrc/internal/logger"
	"golrc/internal/utils"
	"golrc/library"
	"golrc/provider"
	"path"
	"strconv"
	"strings"
)

var (
	log           = logger.NewTaggedLogger("PROVIDER(LRCLIB)")
	lrclibFetcher = fetch.NewFetcher("https://lrclib.net")
)

type Track struct {
	Tag          library.Tag `json:"-"`
	AudioFile    string
	ID           int    `json:"id"`
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func (t Track) GetTag() library.Tag {
	return t.Tag
}

func (t Track) GetAudioFile() string {
	return t.AudioFile
}

func (t Track) HasLyrics() bool {
	return t.PlainLyrics != "" || t.SyncedLyrics != ""
}

func (t Track) GetSyncedLyrics() string {
	if t.SyncedLyrics == "" {
		return ""
	}

	w := fmt.Sprintf(`[ar:%s]
[al:%s]
[ti:%s]
[tool:golrc]
`, t.Tag.Artist, t.Tag.Album, t.Tag.Title)
	return w + t.SyncedLyrics
}

func (t Track) GetPlainLyrics() string {
	return t.PlainLyrics
}

func (t Track) GetLyrics() string {
	if t.GetSyncedLyrics() != "" {
		return t.GetSyncedLyrics()
	}

	return t.GetPlainLyrics()
}

func GetLyrics(tag library.Tag) (Track, error) {
	params := tag.Params()
	params.Set("track_name", tag.Title)
	params.Set("artist_name", tag.Artist)
	params.Set("album_name", tag.Album)
	params.Set("duration", strconv.Itoa(tag.Duration))

	var track Track
	err := lrclibFetcher.Get("/api/get?"+params.Encode(), &track)
	if err != nil {
		log.E("Error fetching lyrics from LRCLib", "error", err)
		return track, err
	}
	track.AudioFile = tag.AudioFile
	log.D("Fetched lyrics from LRCLib", "trackID", track.ID)
	track.Tag = tag
	return track, nil
}

func GetAllLyrics(tags []library.Tag) ([]Track, error) {
	in := make(chan library.Tag)
	go func() {
		for _, tag := range tags {
			in <- tag
		}
		close(in)
	}()

	res := utils.ProcessAndGather(in, func(tag library.Tag) (Track, error) {
		zero := Track{}
		log.I("[~] Fetching lyrics", "title", tag.Title, "artist", tag.Artist, "album", tag.Album)

		t, err := GetLyrics(tag)
		if err != nil {
			log.E("Failed to get lyrics", "title", tag.Title, "artist", tag.Artist, "error", err)
			return zero, err
		}
		log.I("[+] Fetched lyrics", "trackID", t.ID, "title", tag.Title, "artist", tag.Artist, "album", tag.Album)
		return t, nil
	}, max(len(tags)/10, 10))

	return res, nil
}

func WriteLyric(track Track) error {
	if track.AudioFile == "" {
		log.E("Audio file is empty, cannot write lyrics", "trackID", track.ID)
		return errors.New("audio file path is empty")
	}

	parts := strings.Split(track.AudioFile, ".")
	lrcPath := parts[0]
	lrcPath += ".lrc"
	log.D("LRC path", "path", lrcPath)

	if track.GetPlainLyrics() == "" && track.GetSyncedLyrics() == "" {
		log.W("No lyrics available to write", "trackID", track.ID)
		return errors.New("no lyrics available to write")
	}

	if track.SyncedLyrics == "" {
		log.D("Plain lyrics", "lyrics", track.GetPlainLyrics())
		err := provider.WriteLRC(lrcPath, track, track.PlainLyrics)
		if err != nil {
			log.E("Failed to write lyrics to file", "path", lrcPath, "error", err)
			return err
		}
		log.I("[+] Wrote plain lyrics to file", "file", path.Base(track.AudioFile), "lrcpath", path.Base(lrcPath))
		return nil
	}

	log.D("Synced lyrics", "lyrics", track.GetSyncedLyrics())
	err := provider.WriteLRC(lrcPath, track, track.GetSyncedLyrics())
	if err != nil {
		log.E("Failed to write synced lyrics to file", "path", lrcPath, "error", err)
		return err
	}
	log.I("[+] Wrote synced lyrics to file", "file", path.Base(track.AudioFile), "lrcpath", path.Base(lrcPath))
	return nil
}

func WriteLyrics(tracks []Track) int {
	in := make(chan Track)
	go func() {
		for _, t := range tracks {
			if t.GetAudioFile() != "" && t.HasLyrics() {
				in <- t
			} else {
				if t.GetTag().Title != "" && t.GetTag().Artist != "" && t.GetTag().Album != "" {
					log.W("[~] Skipping track without lyrics or audio file", "title", t.Tag.Title, "artist", t.Tag.Artist, "album", t.Tag.Album)
				}
			}
		}
		close(in)
	}()

	type ErrMsg struct {
		Track Track
		Err   error
	}

	wrote := 0
	errs := utils.ProcessAndGather(in, func(track Track) (ErrMsg, error) {
		err := WriteLyric(track)
		if err != nil {
			return ErrMsg{track, err}, err
		}

		wrote++
		return ErrMsg{track, nil}, nil
	}, max(len(tracks)/10, 10))

	for _, err := range errs {
		if err.Err != nil {
			log.I("[-] Failed to write lyrics", "trackID", err.Track.ID, "file", err.Track.AudioFile, "error", err.Err)
		}
	}

	return wrote
}
