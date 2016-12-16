package file

import (
	"bufio"
	"os"
)

const (
	IONBF = iota
	IOFBF
	IOLBF
)

type File interface {
	IsClosed() bool
	Close() error
	Write(p []byte) (nn int, err error)
	WriteString(s string) (nn int, err error)
	Flush() error
	UnreadByte() error
	ReadByte() (c byte, err error)
	Read(p []byte) (n int, err error)
	ReadSlice(delim byte) (line []byte, err error)
	Seek(offset int64, whence int) (n int64, err error)
	Setvbuf(mode int, size int) (err error)
}

func newFile(f *os.File, std bool) File {
	return &file{
		File:  f,
		state: seek,
		r:     bufio.NewReader(f),
		w:     bufio.NewWriter(f),
		std:   std,
	}
}

func newReadOnlyFile(f *os.File, std bool) File {
	return &rofile{
		File: f,
		r:    bufio.NewReader(f),
		std:  std,
	}
}

func newWriteOnlyFile(f *os.File, std bool) File {
	return &wofile{
		File: f,
		w:    bufio.NewWriter(f),
		std:  std,
	}
}

func NewFile(f *os.File, flag int, std bool) File {
	switch flag & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		return newReadOnlyFile(f, std)
	case os.O_WRONLY:
		return newWriteOnlyFile(f, std)
	case os.O_RDWR:
		return newFile(f, std)
	default:
		panic("unreachable")
	}
}

func OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return NewFile(f, flag, false), nil
}
