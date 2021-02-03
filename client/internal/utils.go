package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func IsWhitespaceString(value string) bool {
	return len(strings.TrimSpace(value)) == 0
}

func ReadCloseableBuffer(buffer io.ReadCloser) ([]byte, error) {
	byteBuffer := bytes.Buffer{}
	if _, err := byteBuffer.ReadFrom(buffer); err != nil {
		return nil, err
	}

	if err := buffer.Close(); err != nil {
		return nil, err
	}

	return byteBuffer.Bytes(), nil
}

func BufferToMap(buffer io.ReadCloser) (map[string]interface{}, error) {
	if b, err := ReadCloseableBuffer(buffer); err != nil {
		return nil, err
	} else {
		m := make(map[string]interface{})
		if jsonErr := json.Unmarshal(b, &m); jsonErr != nil {
			fmt.Println(jsonErr.Error())
			return nil, jsonErr
		} else {
			return m, nil
		}
	}
}

func JsonContentType(header *http.Header) bool {
	return strings.Contains(header.Get("Content-Type"), jsonMime)
}
