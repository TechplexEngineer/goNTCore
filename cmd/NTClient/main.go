package main

import (
	"fmt"

	"github.com/technomancers/goNTCore"
)

func main() {
	quit := make(chan bool)
	c, err := goNTCore.NewClient("localhost", "Test Client")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	go func() {
		for l := range c.Log {
			if l.Err != nil {
				fmt.Println(l.Err)
			}
			if l.Message != "" {
				fmt.Println(l.Message)
			}
		}
	}()
	go c.Listen()
	c.StartHandshake()
	<-quit
}
