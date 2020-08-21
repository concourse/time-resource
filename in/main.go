package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	var request models.InRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	command := resource.InCommand{}

	response, err := command.Run(destination, request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "running command:", err.Error())
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(response)
}
