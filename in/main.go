package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	versionTime := request.Version.Time
	if versionTime.IsZero() {
		versionTime = time.Now()
	}

	inVersion := request.Version
	inVersion.Time = versionTime

	metadata := models.Metadata{
		{"time", versionTime.String()},
	}

	if request.Source.Interval != "" {
		metadata = append(metadata, models.MetadataField{"interval", request.Source.Interval})
	}

	if request.Source.Start != "" {
		metadata = append(metadata, models.MetadataField{"start", request.Source.Start})
	}

	if request.Source.Stop != "" {
		metadata = append(metadata, models.MetadataField{"stop", request.Source.Stop})
	}

	if len(request.Source.Days) > 0 {
		metadata = append(metadata, models.MetadataField{"days", strings.Join(request.Source.Days[:], ", ")})
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version:  inVersion,
		Metadata: metadata,
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
