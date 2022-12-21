package gui

import (
	"fmt"
	"liberty-town/node/config"
	"liberty-town/node/gui/gui_interface"
	"liberty-town/node/gui/gui_non_interactive"
)

var GUI gui_interface.GUIInterface

//test
func InitGUI() (err error) {

	GUI, err = gui_non_interactive.CreateGUINonInteractive()

	GUI.Info("GO " + config.NAME)
	GUI.Info(fmt.Sprintf("OS: %s ARCH: %s %d", config.OS, config.ARCHITECTURE, config.CPU_THREADS))
	GUI.Info("VERSION " + config.VERSION_STRING)
	GUI.Info("BUILD_VERSION " + config.BUILD_VERSION)

	return
}
