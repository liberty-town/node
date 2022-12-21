//go:build !js
// +build !js

package config

import (
	"errors"
	"liberty-town/node/config/arguments"
	"os"
	"strconv"
)

var ()

func config_init() (err error) {

	if _, err = os.Stat("_data/"); errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir("_data/", 0755); err != nil {
			return
		}
	}

	if ORIGINAL_PATH, err = os.Getwd(); err != nil {
		return
	}

	if err = os.Chdir("./_data"); err != nil {
		return
	}

	var prefix string
	if arguments.Arguments["--instance"] != nil {
		INSTANCE = arguments.Arguments["--instance"].(string)
		prefix = INSTANCE
	} else {
		prefix = "default"
	}

	if arguments.Arguments["--instance-id"] != nil {
		a := arguments.Arguments["--instance-id"].(string)
		if INSTANCE_ID, err = strconv.Atoi(a); err != nil {
			return
		}
	}
	prefix += "_" + strconv.Itoa(INSTANCE_ID)

	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	prefix += "/" + NETWORK_SELECTED_NAME
	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	if err = os.Chdir("./" + prefix); err != nil {
		return
	}

	return
}
