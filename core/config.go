package din

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

var Config config

type config struct {
	Core struct {
		Addr  string `json:"addr"`
		Debug bool   `json:"debug"`
	} `json:"core"`
}

func (c *config) parseFile(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	if err := json.NewDecoder(fi).Decode(c); err != nil {
		return err
	}
	return nil
}

func (c *config) dump(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

func (c *config) dumpToFile(path string, overWrite bool) (err error) {
	var flag int
	if overWrite {
		flag = os.O_CREATE | os.O_RDWR | os.O_TRUNC
	} else {
		flag = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	}
	fi, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return err
	}
	defer fi.Close()

	return c.dump(fi)
}

func ParseConfigFile(path string) error {
	return Config.parseFile(path)
}

// locateConfig looks for the din configuration file.  The configuration file
// is looked for in the following locations:
//
//   - environment variable DIN_CONFIG
//   - file named config.json in the current directory
//
func locateConfig() string {
	path := os.Getenv("DIN_CONFIG")
	if path != "" {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "config.json")
}
