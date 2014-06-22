package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/concourse/time-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	err := os.MkdirAll(destination, 0755)
	if err != nil {
		fatal("creating destination", err)
	}

	file, err := os.Create(filepath.Join(destination, "input"))
	if err != nil {
		fatal("creating input file", err)
	}

	defer file.Close()

	var request models.InRequest

	err = json.NewDecoder(io.TeeReader(os.Stdin, file)).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version: models.Version{Time: time.Now()},
		Metadata: models.Metadata{
			{"interval", request.Source.Interval},
			{"now", time.Now().String()},
		},
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
