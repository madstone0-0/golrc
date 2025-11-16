package utils

import (
	"golrc/internal/logger"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var (
	LocalEnv = map[string]string{}
	log      = logger.NewTaggedLogger("INTERNAL:(UTILS)")
)

func ResolvePath(path string) (absPath string, err error) {
	user, err := user.Current()
	if err != nil {
		return absPath, err
	}
	homeDir := user.HomeDir
	absPath = filepath.Clean(path)
	absPath = strings.Replace(absPath, "~", homeDir, 1)

	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return absPath, err
	}

	if !filepath.IsAbs(absPath) {
		return absPath, err
	}

	return absPath, nil
}

func OrDefault[T any](value T, defaultValue ...T) T {
	if reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
		for _, def := range defaultValue {
			if !reflect.DeepEqual(def, reflect.Zero(reflect.TypeOf(def)).Interface()) {
				return def
			}
		}
	}
	return value
}

func Env(key string) string {
	v, ok := LocalEnv[key]
	if !ok {
		v, ok = os.LookupEnv(key)
		if !ok {
			log.F("Environment variable not found", "key", key)
		}
	}
	return v
}

func EnvCanFail(key string) string {
	v, ok := LocalEnv[key]
	if !ok {
		v, ok = os.LookupEnv(key)
		if !ok {
			log.D("Environment variable not found or empty", "key", key)
			return ""
		}
	}
	return v
}

func Write(path string, data string) error {
	return os.WriteFile(path, []byte(data), os.ModePerm)
}

func ProcessAndGather[T, R any](in <-chan T, processor func(T) (R, error), num int) []R {
	out := make(chan R, num)
	var wg sync.WaitGroup
	wg.Add(num)

	// Process
	for range num {
		go func() {
			defer wg.Done()
			for v := range in {
				// Gather in out
				val, err := processor(v)
				if err != nil {
					log.E("Error processing value", "error", err)
				}
				out <- val
			}
		}()
	}

	// Wait until all sender routines are done before closing output channel
	go func() {
		wg.Wait()
		close(out)
	}()

	var res []R
	for v := range out {
		res = append(res, v)
	}
	return res
}
