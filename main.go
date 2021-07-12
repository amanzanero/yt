package main

import (
	"github.com/amanzanero/yt/cmd"
)

func main() {
	executeErr := cmd.Execute()
	if executeErr != nil {
		return
	}
}
