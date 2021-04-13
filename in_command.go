package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/concourse/time-resource/models"
)

type InCommand struct {
}

func (*InCommand) Run(destination string, request models.InRequest) (models.InResponse, error) {
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating destination: %w", err)
	}

	file, err := os.Create(filepath.Join(destination, "input"))
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating input file: %w", err)
	}

	defer file.Close()

	err = json.NewEncoder(file).Encode(request)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("writing input file: %w", err)
	}

	versionTime := request.Version.Time
	if versionTime.IsZero() {
		versionTime = time.Now()
	}

	timeFile, err := os.Create(filepath.Join(destination, "timestamp"))
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating input file: %w", err)
	}
	timeFile.WriteString(versionTime.Format("2006-01-02 15:04:05.999999999 -0700 MST"))

	inVersion := models.Version{Time: versionTime}
	response := models.InResponse{Version: inVersion}

	return response, nil
}
