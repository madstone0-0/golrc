package appstate

import (
	"golrc/internal/utils"
	"os"
)

var (
	HOME           = utils.OrDefault(os.Getenv("HOME"), os.Getenv("USERPROFILE"), "/tmp")
	XDG_CACHE_DIR  = utils.OrDefault(os.Getenv("XDG_CACHE_HOME"), HOME+"/.cache/")
	XDG_CONFIG_DIR = utils.OrDefault(os.Getenv("XDG_CONFIG_HOME"), HOME+"/.config/")
)
