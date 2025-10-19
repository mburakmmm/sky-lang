package skylib

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// TOMLParse parses TOML string
func TOMLParse(data string) (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := toml.Decode(data, &result)
	return result, err
}

// TOMLEncode encodes to TOML
func TOMLEncode(obj interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	err := encoder.Encode(obj)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
