package main

import (
	"fmt"

	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, string](10)

	value, status := mp.Getset("exampleKey", func() string {
		return "This is a computed value"
	})
	fmt.Println("Status:", status, "Value:", value)

	value, status = mp.Getset("exampleKey", func() string {
		return "New value"
	})
	fmt.Println("Status:", status, "Value:", value)
}
