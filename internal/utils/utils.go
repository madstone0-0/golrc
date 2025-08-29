package utils

import (
	"golrc/internal/logger"
	"os"
	"reflect"
)

var (
	LocalEnv = map[string]string{}
	log      = logger.NewTaggedLogger("INTERNAL:(UTILS)")
)

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
