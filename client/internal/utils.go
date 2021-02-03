package internal

import (
	"bytes"
	"io"
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
