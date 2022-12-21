package recovery

import (
	"fmt"
	"os"
	"runtime/debug"
)

func SafeGo(cb func()) {
	go func() {
		Safe(cb)
	}()
}

func Safe(cb func()) {
	defer func() {
		if err := recover(); err != nil {

			stackTrace := string(debug.Stack())

			fmt.Println("Error: \n\n", err, stackTrace)
			os.Exit(1)
		}
	}()
	cb()
}
