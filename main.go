package main

import (
	"liberty-town/node/start"
)

func main() {
	if err := start.InitMain(nil); err != nil {
		panic(err)
	}
}
