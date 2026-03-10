package main

import (
	"fmt"

	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, int](10)

	mp.Set("key1", 100)
	mp.Set("key2", 200)

	if value, found := mp.Get("key1"); found {
		fmt.Println("Key1 Value:", value)
	}

	mp.Range(func(key string, value int) bool {
		fmt.Println(key, value)
		return true
	})
}
