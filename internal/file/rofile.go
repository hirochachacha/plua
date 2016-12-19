package file

import (
	"bufio"
	"errors"
	"io"
	"os"
)

type rofile struct {
	*os.File
	br     *bufio.Reader
	off    int64
	closed bool
	std    bool
}

func (ro *rofile) IsClosed() bool {
	return ro.closed
}

func (ro *rofile) Close() error {
	if ro.std {
		return errors.New("cannot close standard file")
	}

	if err := ro.File.Close(); err != nil {
		return err
	}

	ro.closed = true

	return nil
}

func (ro *rofile) Flush() error {
	return os.ErrInvalid
}

func (ro *rofile) UnreadByte() error {
	err := ro.br.UnreadByte()
	if err == nil {
		ro.off--
	}
	return err
}

func (ro *rofile) ReadByte() (c byte, err error) {
	c, err = ro.br.ReadByte()
	if err == nil {
		ro.off++
	}
	return c, err
}

func (ro *rofile) Read(p []byte) (n int, err error) {
	n, err = ro.br.Read(p)
	ro.off += int64(n)
	return n, err
}

func (ro *rofile) ReadBytes(delim byte) (line []byte, err error) {
	line, err = ro.br.ReadBytes(delim)
	ro.off += int64(len(line))
	return line, err
}

func (ro *rofile) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekStart:
		if ro.off <= offset && offset <= ro.off+int64(ro.br.Buffered()) {
			ro.br.Discard(int(offset - ro.off))
			ro.off = offset
		} else {
			ro.off, err = ro.File.Seek(offset, io.SeekStart)
			ro.br.Reset(ro.File)
		}
	case io.SeekCurrent:
		if 0 <= offset && offset <= int64(ro.br.Buffered()) {
			ro.br.Discard(int(offset))
			ro.off += offset
		} else {
			ro.off, err = ro.File.Seek(ro.off+offset, io.SeekStart)
			ro.br.Reset(ro.File)
		}
	case io.SeekEnd:
		ro.off, err = ro.File.Seek(offset, io.SeekEnd)
		ro.br.Reset(ro.File)
	}

	n = ro.off

	return
}

func (ro *rofile) Setvbuf(mode int, size int) (err error) {
	_, err = ro.File.Seek(ro.off, io.SeekStart)

	if size > 0 {
		ro.br = bufio.NewReaderSize(ro.File, size)
	}

	return
}
