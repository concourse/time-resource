package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/models"
)

func main() {
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	command := resource.CheckCommand{}

	versions, err := command.Run(request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "running command:", err.Error())
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
