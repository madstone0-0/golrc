package provider

import (
	"errors"
	"golrc/internal/appstate"
	"golrc/internal/dts/heap"
	"golrc/internal/logger"
	"golrc/internal/utils"
	"golrc/library"
	"path"
	"slices"
)

var (
	log = logger.NewTaggedLogger("PROVIDER")
)

type ITrack interface {
	GetTag() library.Tag
	GetAudioFile() string
	GetLyrics() string
	GetPlainLyrics() string
	GetSyncedLyrics() string
	HasLyrics() bool
	GetId() int
}

func WriteLRC(p string, track ITrack, data string) error {
	if appstate.DRY_RUN {
		log.I("[+] Would write lyrics", "file", path.Base(p), "track", path.Base(track.GetAudioFile()))
		return nil
	}
	return utils.Write(p, data)
}

type Provider interface {
	GetName() string
	GetAndWriteLyrics([]library.Tag) (int, []library.Tag, error)
}

type ProviderCtor struct {
	Priority int
	Name     string
	Ctor     func() (Provider, error)
}

type ProviderItem struct {
	Priority int
	Provider Provider
}

func cmp(a, b ProviderItem) int {
	// Higher priority first
	return slices.Compare([]int{b.Priority}, []int{a.Priority})
}

func CreateProviders(ctors ...ProviderCtor) (*heap.Heap[ProviderItem], error) {
	providers := heap.New[ProviderItem](cmp)
	for _, ctor := range ctors {
		prv, err := ctor.Ctor()
		if err != nil {
			log.E("[-] Could not create provider", "provider", ctor.Name, "error", err)
			continue
		}
		providers.Push(ProviderItem{
			Priority: ctor.Priority,
			Provider: prv,
		})
		log.I("[+] Created provider", "provider", ctor.Name)
	}

	if providers.Size() == 0 {
		return nil, errors.New("no providers could be created")
	}
	return providers, nil
}

func TryProvider(tags []library.Tag, prv Provider) error {
	wrote, _, err := prv.GetAndWriteLyrics(tags)
	if err != nil {
		log.E("[-] Failed to get and write lyrics", "provider", prv.GetName(), "error", err)
		return err
	}

	log.I("[+] Lyrics fetching completed", "total", len(tags), "written", wrote)
	return nil
}

func TryProviders(tags []library.Tag, ctors ...ProviderCtor) error {
	providers, err := CreateProviders(ctors...)
	if err != nil {
		return err
	}

	searchTags := tags
	for providers.Size() > 0 {
		prvItem, ok := providers.Pop()
		if !ok {
			break
		}

		prv := prvItem.Provider
		log.I("[+] Trying provider", "provider", prv.GetName())
		log.D("[*] Provider details", "provider", prv.GetName(), "remaining_tags", len(searchTags), "priority", prvItem.Priority)
		wrote, notFound, err := prv.GetAndWriteLyrics(searchTags)
		searchTags = notFound
		if err != nil {
			log.E("[-] Failed to get and write lyrics", "provider", prv.GetName(), "error", err)
			continue
		}

		log.I("[+] Lyrics fetching completed", "total", len(tags), "written", wrote)
	}
	return nil
}
