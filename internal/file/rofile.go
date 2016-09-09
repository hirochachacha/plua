package file

import (
	"bufio"
	"os"
)

type rofile struct {
	*os.File
	r      *bufio.Reader
	closed bool
}

func (ro *rofile) IsClosed() bool {
	return ro.closed
}

func (ro *rofile) Close() error {
	err := ro.File.Close()

	ro.closed = true

	return err
}

func (ro *rofile) Flush() error {
	return os.ErrInvalid
}

func (ro *rofile) UnreadByte() error {
	return ro.r.UnreadByte()
}

func (ro *rofile) ReadByte() (c byte, err error) {
	return ro.r.ReadByte()
}

func (ro *rofile) Read(p []byte) (n int, err error) {
	return ro.r.Read(p)
}

func (ro *rofile) ReadSlice(delim byte) (line []byte, err error) {
	return ro.r.ReadSlice(delim)
}

func (ro *rofile) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case 0:
		n, err = ro.File.Seek(offset, 0)
	case 1:
		n, err = ro.File.Seek(offset-int64(ro.r.Buffered()), 1)
		ro.r.Reset(ro.File)
	case 2:
		n, err = ro.File.Seek(offset, 2)
	}

	return
}

func (ro *rofile) Setvbuf(mode int, size int) (err error) {
	_, err = ro.Seek(-int64(ro.r.Buffered()), 1)
	if err != nil {
		return
	}

	if size > 0 {
		ro.r = bufio.NewReaderSize(ro.File, size)
	}

	return
}
