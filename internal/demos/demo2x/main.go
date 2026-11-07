package main

import (
	"fmt"

	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, string](10)

	// First time: the value is absent, so it is computed and stored, loaded is false
	value, loaded := mp.Getset("exampleID", func() string {
		return "This is a computed value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)

	// Second time: the value is in the map, the function does not run, loaded is true
	value, loaded = mp.Getset("exampleID", func() string {
		return "New value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)
}
