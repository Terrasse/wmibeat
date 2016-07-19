package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/eskibars/wmibeat/beater"
)

var Name = "wmibeat"

func main() {
	if err := beat.Run(Name, "", beater.New); err != nil {
		os.Exit(1)
	}
}
