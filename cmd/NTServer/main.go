// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/technomancers/goNTCore"
)

func main() {
	s, err := goNTCore.NewServer("Test", nil)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	s.StartPeriodicClean(5 * time.Second)
	s.Listen()
}
