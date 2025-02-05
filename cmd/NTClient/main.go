// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/techplexengineer/gontcore"
	"github.com/techplexengineer/gontcore/storage"
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

	//c.Debug = true

	c.AddKeyListener("newEntry", func(entry storage.StorageEntry) {
		log.Printf("Listener: %#v", entry)
	})

	log.Printf("--- Client Started ---\n\n")
	defer c.Close()

	time.Sleep(2 * time.Second)

	//c.PutBoolean("newEntry", true)
	//
	//time.Sleep(2 * time.Second)
	//
	//c.PutBoolean("newEntry", false)

	snap := c.GetSnapshot()
	data, err := json.MarshalIndent(snap, "", "    ")
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	log.Printf("%s", data)

	key := "/Shuffleboard/Robot Setup/Encoder LR/Zero!"
	res := c.GetBoolean(key, false)
	log.Printf("%s - %v", key, res)

	key = "/FMSInfo/IsRedAlliance"
	res = c.GetBoolean(key, false)
	log.Printf("%s - %v", key, res)

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
