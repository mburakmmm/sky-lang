package skylib

import (
	"archive/zip"
	"bytes"
	"io"
)

// ZipCompress creates a zip archive
func ZipCompress(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for name, data := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			return nil, err
		}

		_, err = fileWriter.Write(data)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ZipDecompress extracts a zip archive
func ZipDecompress(data []byte) (map[string][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte)

	for _, file := range reader.File {
		fileReader, err := file.Open()
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			return nil, err
		}

		files[file.Name] = content
	}

	return files, nil
}
