package client

import (
	"io"
	"io/ioutil"
)

func ReadCloseableBuffer(buffer io.ReadCloser) ([]byte, error) {
	// close errors are ignored, since all data should be available
	defer buffer.Close()

	buf, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
