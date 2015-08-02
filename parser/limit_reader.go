package parser

import (
	"io"
	"io/ioutil"
)

type limitReader struct {
	io.Reader
}

func newLimitReader(r io.Reader, limit int64) *limitReader {
	return &limitReader{
		Reader: io.LimitReader(r, limit),
	}
}

func (r *limitReader) Close() error {
	if _, err := io.Copy(ioutil.Discard, r); err != nil {
		return err
	}
	return nil
}
