package main

import (
	"log"

	"github.com/judah-caruso/brutengine/engine"
)

func main() {
	err := engine.Setup()
	if err != nil {
		log.Fatal(err)
	}

	defer engine.Teardown()

	err = engine.Run()
	if err != nil {
		log.Fatal(err)
	}
}
