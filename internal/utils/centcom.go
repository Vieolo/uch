package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vieolo/filange"
)

type CentComCommand struct {
	Cmd string `json:"cmd"`
}

type CentComAll struct {
	Commands map[string]CentComCommand `json:"commands"`
}

type CentComPath struct {
	ConfigDir      string
	ConfigFileName string
	FullPath       string
}

func GetCentComPath() (CentComPath, error) {
	configFolder, err := os.UserHomeDir()
	if err != nil {
		return CentComPath{}, fmt.Errorf("error while getting the config path: %w", err)
	}
	cd := fmt.Sprintf("%s/.uch", configFolder)
	return CentComPath{
		ConfigDir:      cd,
		ConfigFileName: "config.json",
		FullPath:       fmt.Sprintf("%s/%s", cd, "config.json"),
	}, nil
}

// Gets or create config.json file
func GetCentCom() (CentComAll, error) {
	centPath, err := GetCentComPath()
	if err != nil {
		return CentComAll{}, err
	}

	exists := filange.FileExists(centPath.FullPath)
	if !exists {
		if err := filange.CreateDirIfNotExists(centPath.ConfigDir, os.FileMode(0700)); err != nil {
			return CentComAll{}, err
		}
		if err := os.WriteFile(centPath.FullPath, []byte("{\"commands\":{}}"), os.FileMode(0600)); err != nil {
			return CentComAll{}, err
		}
	}

	centBytes, err := os.ReadFile(centPath.FullPath)
	if err != nil {
		return CentComAll{}, err
	}

	var cc CentComAll
	if err = json.Unmarshal(centBytes, &cc); err != nil {
		return CentComAll{}, fmt.Errorf("Malformed config.json file: %w", err)
	}

	return cc, nil
}
