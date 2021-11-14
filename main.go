// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"os"

	"jdk.sh/meta"
)

func main() {
	if err := mainCmd(); err != nil {
		log.Println("joshdk/drone-skip-pipeline:", err)
		os.Exit(1)
	}
}

func mainCmd() error {
	log.Printf("joshdk/drone-skip-pipeline %s (%s)\n", meta.Version(), meta.ShortSHA())
	return nil
}
