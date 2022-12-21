package app

import (
	"liberty-town/node/gui"
	"liberty-town/node/store"
)

func Close() (err error) {
	if err = store.DBClose(); err != nil {
		return
	}
	gui.GUI.Close()
	return nil
}
