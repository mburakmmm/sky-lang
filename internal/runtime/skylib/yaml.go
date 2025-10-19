package skylib

import (
	"gopkg.in/yaml.v3"
)

// YAMLParse parses YAML string
func YAMLParse(data string) (interface{}, error) {
	var result interface{}
	err := yaml.Unmarshal([]byte(data), &result)
	return result, err
}

// YAMLEncode encodes to YAML
func YAMLEncode(obj interface{}) (string, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

