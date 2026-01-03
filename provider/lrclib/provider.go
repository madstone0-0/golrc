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
	"sync"
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
	Intrumental  bool   `json:"instrumental"`
}

func (t Track) GetTag() library.Tag {
	return t.Tag
}

func (t Track) GetAudioFile() string {
	return t.AudioFile
}

func (t Track) HasLyrics() bool {
	return t.Intrumental || (t.PlainLyrics != "" || t.SyncedLyrics != "")
}

func (t Track) GetSyncedLyrics() string {
	if t.Intrumental {
		w := fmt.Sprintf(`[ar:%s]
[al:%s]
[ti:%s]
[tool:golrc]
[00:00.00]♪ Instrumental ♪
`, t.Tag.Artist, t.Tag.Album, t.Tag.Title)
		return w
	}

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

func (t Track) GetId() int {
	return t.ID
}

type LrcLib struct {
}

func getLyrics(tag library.Tag) (Track, error) {
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
	log.D("Track", "track", track)
	track.AudioFile = tag.AudioFile
	log.D("Fetched lyrics from LRCLib", "trackID", track.ID)
	track.Tag = tag
	return track, nil
}

func (l LrcLib) GetAllLyrics(tags []library.Tag) ([]Track, []library.Tag, error) {
	notFound := make([]library.Tag, 0, len(tags))
	nFChan := make(chan library.Tag)
	in := make(chan library.Tag)
	go func() {
		for _, tag := range tags {
			in <- tag
		}
		close(in)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for tag := range nFChan {
			notFound = append(notFound, tag)
		}
	}()

	res := utils.ProcessAndGather(in, func(tag library.Tag) (Track, error) {
		zero := Track{}
		log.I("[~] Fetching lyrics", "title", tag.Title, "artist", tag.Artist, "album", tag.Album)

		t, err := getLyrics(tag)
		if err != nil {
			log.E("Failed to get lyrics", "title", tag.Title, "artist", tag.Artist, "error", err)
			nFChan <- tag
			return zero, err
		}
		log.I("[+] Fetched lyrics", "trackID", t.ID, "title", tag.Title, "artist", tag.Artist, "album", tag.Album)
		return t, nil
	}, max(len(tags)/10, 10))

	close(nFChan)
	wg.Wait()

	return res, notFound, nil
}

func writeLyric(track Track) error {
	if track.AudioFile == "" {
		log.E("Audio file is empty, cannot write lyrics", "trackID", track.ID)
		return errors.New("audio file path is empty")
	}

	lrcPath := strings.TrimSuffix(track.AudioFile, path.Ext(track.AudioFile))
	lrcPath += ".lrc"
	log.D("LRC path", "path", lrcPath)

	if track.GetPlainLyrics() == "" && track.GetSyncedLyrics() == "" {
		log.W("No lyrics available to write", "trackID", track.ID)
		return errors.New("no lyrics available to write")
	}

	if !track.Intrumental && track.SyncedLyrics == "" {
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

func (l LrcLib) WriteLyrics(tracks []Track) int {
	in := make(chan Track)
	go func() {
		for _, t := range tracks {
			if t.GetAudioFile() != "" && t.HasLyrics() {
				in <- t
			} else {
				if t.GetTag().Title != "" && t.GetTag().Artist != "" && t.GetTag().Album != "" {
					log.W("[~] Skipping track without lyrics or audio file", "title", t.GetTag().Title, "artist", t.GetTag().Artist, "album", t.GetTag().Album)
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
		err := writeLyric(track)
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

func (l LrcLib) GetAndWriteLyrics(tags []library.Tag) (int, []library.Tag, error) {
	tracks, notFound, err := l.GetAllLyrics(tags)
	if err != nil {
		log.E("Failed to get lyrics", "error", err)
		return 0, nil, err
	}
	log.I("[+] Fetched lyrics", "count", len(tracks))

	wrote := l.WriteLyrics(tracks)
	return wrote, notFound, nil
}

func (l LrcLib) GetName() string {
	return "LrcLib"
}

func New() (LrcLib, error) {
	return LrcLib{}, nil
}
