package library

import (
	"errors"
	"golrc/internal/logger"
	"golrc/internal/utils"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"go.senan.xyz/taglib"
)

var (
	log = logger.NewTaggedLogger("TAG")
)

var MUSIC_EXT = []string{"mp3", "flac", "m4a", "wav", "ogg", "opus", "aac", "wma"}

func ReadMusicDir(dir string, filterExisting bool) ([]string, error) {
	var musicFiles []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != "" {
			ext = ext[1:] // remove the dot
		}

		if ext == "" || !slices.Contains(MUSIC_EXT, strings.ToLower(ext)) {
			return nil
		}

		lrcFile := strings.TrimSuffix(path, filepath.Ext(path)) + ".lrc"
		info, err := os.Stat(lrcFile)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if filterExisting {
			if err == nil && info.Size() > 0 {
				log.D("[+] Found associated LRC file. Skipping...", "file", lrcFile)
				return nil
			}
		}

		musicFiles = append(musicFiles, path)
		log.D("[+] Found music file", "file", path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return musicFiles, nil
}

type Tag struct {
	AudioFile string
	Title     string `json:"track_name"`
	Artist    string `json:"artist_name"`
	Album     string `json:"album_name"`
	Duration  int    `json:"duration"`
}

func (t Tag) IsValid() bool {
	return t.Title != "" && t.Artist != ""
}

func (t Tag) Params() url.Values {
	fv := reflect.ValueOf(t)
	ft := reflect.TypeOf(t)

	params := url.Values{}
	for i := 0; i < ft.NumField(); i++ {
		field := ft.Field(i)
		value := fv.Field(i).Interface()

		strValue := ""
		switch v := value.(type) {
		case string:
			strValue = v
		case int:
			strValue = strconv.Itoa(v)
		}

		if strValue != "" {
			params.Add(field.Tag.Get("json"), strValue)
		}
	}
	return params
}

func ParseTags(filepaths []string) ([]Tag, error) {
	in := make(chan string)
	go func() {
		for _, path := range filepaths {
			in <- path
		}
		close(in)
	}()

	metadata := utils.ProcessAndGather(in, func(path string) (Tag, error) {
		zero := Tag{}

		m, err := taglib.ReadTags(path)
		if err != nil {
			log.E("Error parsing tags", "file", path, "error", err)
			return zero, err
		}

		p, err := taglib.ReadProperties(path)
		if err != nil {
			log.E("Error reading properties", "file", path, "error", err)
			return zero, err
		}

		duration := int(p.Length.Seconds())
		tag := Tag{
			AudioFile: path,
			Title:     m[taglib.Title][0],
			Artist:    m[taglib.Artist][0],
			Album:     m[taglib.Album][0],
			Duration:  duration,
		}
		log.I("[+] Parsed tags", "title", tag.Title, "artist", tag.Artist, "album", tag.Album)
		return tag, err
	}, max(len(filepaths)/10, 10))

	return metadata, nil
}
