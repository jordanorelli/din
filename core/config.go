package din

import (
	"encoding/json"
	"os"
)

var Config config

type config struct {
	Core struct {
		Addr  string `json:"addr"`
		Debug bool   `json:"debug"`
	} `json:"core"`
}

func ParseConfigFile(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	if err := json.NewDecoder(fi).Decode(&Config); err != nil {
		return err
	}
	return nil
}
