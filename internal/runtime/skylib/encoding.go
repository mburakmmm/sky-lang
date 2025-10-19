package skylib

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"bytes"
	"io"
	"strings"
)

// JSON encoding

// JSONEncode encodes to JSON
func JSONEncode(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// JSONDecode decodes from JSON
func JSONDecode(data string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}

// JSONEncodePretty encodes to pretty JSON
func JSONEncodePretty(obj interface{}) (string, error) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CSV encoding

// CSVParse parses CSV string
func CSVParse(data string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(data))
	return reader.ReadAll()
}

// CSVEncode encodes to CSV
func CSVEncode(rows [][]string) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	err := writer.WriteAll(rows)
	if err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// Compression

// GzipCompress compresses data with gzip
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GzipDecompress decompresses gzip data
func GzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	return io.ReadAll(reader)
}

