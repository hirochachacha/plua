package file

import (
	"bufio"
	"os"
)

type wmode int

const (
	wno wmode = iota
	wfull
	wline
)

type iFile interface {
	Close() error
	Write(p []byte) (nn int, err error)
	WriteString(s string) (nn int, err error)
	Flush() error
	UnreadByte() error
	ReadByte() (c byte, err error)
	Read(p []byte) (n int, err error)
	ReadSlice(delim byte) (line []byte, err error)
	Seek(offset int64, whence int) (n int64, err error)
	SetVBuf(wmode wmode, size int) (err error)
}

func newFile(f *os.File) iFile {
	return &file{
		f:     f,
		state: seek,
		r:     bufio.NewReader(f),
		w:     bufio.NewWriter(f),
	}
}

func newReadOnlyFile(f *os.File) iFile {
	return &rofile{
		f: f,
		r: bufio.NewReader(f),
	}
}

func newReadOnlyFileSize(f *os.File, size int) iFile {
	return &rofile{
		f: f,
		r: bufio.NewReaderSize(f, size),
	}
}

func newWriteOnlyFile(f *os.File) iFile {
	return &wofile{
		f: f,
		w: bufio.NewWriter(f),
	}
}

func newWriteOnlyFileMode(f *os.File, wmode wmode) iFile {
	return &wofile{
		f:     f,
		w:     bufio.NewWriter(f),
		wmode: wmode,
	}
}

func newWriteOnlyFileModeSize(f *os.File, wmode wmode, size int) iFile {
	return &wofile{
		f:     f,
		w:     bufio.NewWriterSize(f, size),
		wmode: wmode,
	}
}

func openFile(name string, flag int, perm os.FileMode) (file iFile, err error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	switch {
	case flag&os.O_RDONLY != 0:
		file = newReadOnlyFile(f)
	case flag&os.O_WRONLY != 0:
		file = newWriteOnlyFile(f)
	case flag&os.O_RDWR != 0:
		file = newFile(f)
	default:
		panic("unreachable")
	}

	return file, nil
}
