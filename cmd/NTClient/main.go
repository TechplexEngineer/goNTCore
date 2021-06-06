// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/technomancers/goNTCore"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	c, errorChan, err := goNTCore.NewClient("localhost", "Test Client", nil)
	if err != nil {
		panic(err)
	}

	log.Printf("--- Client Started ---\n\n")
	defer c.Close()

	time.Sleep(2 * time.Second)

	c.PutBoolean("newEntry", true)

	time.Sleep(2 * time.Second)

	c.PutBoolean("newEntry", false)

	ctrlc := make(chan os.Signal)
	signal.Notify(ctrlc, os.Interrupt)

	select {
	case <-ctrlc:
		err = c.Close()
		log.Print("--- Client Stopped by interrupt signal ---")
		if err != nil {
			log.Printf("Client errored on stop: %s", err)
		}
	case err = <-errorChan:
		log.Fatalf("Client exited unexpectedly: %s", err)
		// @todo retry the connection...
	}
}
