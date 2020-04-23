package config

import (
	"encoding/json"
	"os"
)

func ParseJsonConfig(filename string, config interface{}) error {
	configFile, err := os.Open(filename)
	defer func() {
		_ = configFile.Close()
	}()
	if err != nil {
		return err
	}
	err = json.NewDecoder(configFile).Decode(config)
	if err != nil {
		return err
	}
	return nil
}
