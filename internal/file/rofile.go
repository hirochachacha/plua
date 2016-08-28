package file

import (
	"bufio"
	"os"
)

type rofile struct {
	f *os.File
	r *bufio.Reader
}

func (ro *rofile) Close() error {
	return ro.f.Close()
}

func (ro *rofile) Write(p []byte) (nn int, err error) {
	return 0, ErrInvalid
}

func (ro *rofile) WriteString(s string) (int, error) {
	return 0, ErrInvalid
}

func (ro *rofile) Flush() error {
	return ErrInvalid
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
		n, err = ro.f.Seek(offset, 0)
	case 1:
		n, err = ro.f.Seek(offset-int64(ro.r.Buffered()), 1)
		ro.r.Reset(ro.f)
	case 2:
		n, err = ro.f.Seek(offset, 2)
	}

	return
}

func (ro *rofile) SetVBuf(wmode wmode, size int) (err error) {
	_, err = ro.Seek(-int64(ro.r.Buffered()), 1)
	if err != nil {
		return
	}

	if size > 0 {
		ro.r = bufio.NewReaderSize(ro.f, size)
	}

	return
}
