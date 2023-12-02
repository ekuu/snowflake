package file

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type storage struct {
	path string
	file *os.File
}

func NewStorage(path string) (*storage, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &storage{file: file, path: path}, nil
}

func (s *storage) Get() (t int64, err error) {
	b, err := io.ReadAll(s.file)
	if err != nil {
		return t, err
	}
	c := string(b)
	if c == "" {
		return t, fmt.Errorf("the file(%s) storing time information is empty", s.path)
	}
	return strconv.ParseInt(c, 10, 64)
}

func (s *storage) Save(t int64) error {
	_, err := s.file.WriteAt([]byte(strconv.FormatInt(t, 10)), 0)
	return err
}
