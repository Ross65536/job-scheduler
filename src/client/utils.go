package client

import (
	"io"
	"io/ioutil"
	"strings"
)

func IsWhitespaceString(value string) bool {
	return len(strings.TrimSpace(value)) == 0
}

func ReadCloseableBuffer(buffer io.ReadCloser) ([]byte, error) {
	// close errors are ignored, since all data should be available
	defer buffer.Close()

	buf, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
