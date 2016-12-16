package file

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
)

type state int

const (
	seek state = iota
	read
	write
)

type file struct {
	*os.File
	mode   int
	state  state // previous action type
	r      *bufio.Reader
	w      *bufio.Writer
	closed bool
	std    bool
}

func (f *file) IsClosed() bool {
	return f.closed
}

func (f *file) Close() error {
	if f.std {
		return errors.New("cannot close standard file")
	}

	err := f.File.Close()

	f.closed = true

	return err
}

func (f *file) Write(p []byte) (nn int, err error) {
	if f.state == read {
		_, err = f.Seek(-int64(f.r.Buffered()), 1)
		if err != nil {
			return
		}

		f.r.Reset(f.File)
	}

	switch f.mode {
	case IONBF:
		nn, err = f.File.Write(p)
		if err != nil {
			return
		}
	case IOFBF:
		nn, err = f.w.Write(p)
		if err != nil {
			return
		}
	case IOLBF:
		i := bytes.LastIndex(p, []byte{'\n'})
		if i == -1 {
			nn, err = f.w.Write(p)
			if err != nil {
				return
			}
		} else {
			nn, err = f.w.Write(p[:i])
			if err != nil {
				return
			}
			err = f.w.Flush()
			if err != nil {
				return
			}

			nn, err = f.w.Write(p[i+1:])
			if err != nil {
				return
			}
		}
	}

	f.state = write

	return
}

func (f *file) WriteString(s string) (nn int, err error) {
	if f.state == read {
		_, err = f.Seek(-int64(f.r.Buffered()), 1)
		if err != nil {
			return
		}

		f.r.Reset(f.File)
	}

	switch f.mode {
	case IONBF:
		nn, err = f.File.WriteString(s)
		if err != nil {
			return
		}
	case IOFBF:
		nn, err = f.w.WriteString(s)
		if err != nil {
			return
		}
	case IOLBF:
		i := strings.LastIndex(s, "\n")
		if i == -1 {
			nn, err = f.w.WriteString(s)
			if err != nil {
				return
			}
		} else {
			nn, err = f.w.WriteString(s[:i])
			if err != nil {
				return
			}
			err = f.w.Flush()
			if err != nil {
				return
			}

			nn, err = f.w.WriteString(s[i+1:])
			if err != nil {
				return
			}
		}
	}

	f.state = write

	return
}

func (f *file) Flush() error {
	return f.w.Flush()
}

func (f *file) UnreadByte() (err error) {
	if f.state == write {
		err = f.w.Flush()
		if err != nil {
			return
		}
	}

	err = f.r.UnreadByte()

	f.state = read

	return
}

func (f *file) ReadByte() (c byte, err error) {
	if f.state == write {
		err = f.w.Flush()
		if err != nil {
			return
		}
	}

	c, err = f.r.ReadByte()
	if err != nil {
		return
	}

	f.state = read

	return
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.state == write {
		err = f.w.Flush()
		if err != nil {
			return
		}
	}

	n, err = f.r.Read(p)
	if err != nil {
		return
	}

	f.state = read

	return
}

func (f *file) ReadSlice(delim byte) (line []byte, err error) {
	if f.state == write {
		err = f.w.Flush()
		if err != nil {
			return
		}

		f.r.Reset(f.File)
	}

	line, err = f.r.ReadSlice(delim)
	if err != nil {
		return
	}

	f.state = read

	return
}

func (f *file) Seek(offset int64, whence int) (n int64, err error) {
	if f.state == write {
		err = f.w.Flush()
		if err != nil {
			return
		}
	}

	switch whence {
	case 0:
		n, err = f.File.Seek(offset, 0)
	case 1:
		if f.state == read {
			n, err = f.File.Seek(offset-int64(f.r.Buffered()), 1)
			f.r.Reset(f.File)
		} else {
			n, err = f.File.Seek(offset, 1)
		}
	case 2:
		n, err = f.File.Seek(offset, 2)
	}

	if err != nil {
		return
	}

	f.state = seek

	return
}

func (f *file) Setvbuf(mode int, size int) (err error) {
	switch f.state {
	case read:
		_, err = f.Seek(-int64(f.r.Buffered()), 1)
		if err != nil {
			return err
		}
	case write:
		err = f.w.Flush()
		if err != nil {
			return err
		}
	}

	f.mode = mode
	f.state = seek

	if size > 0 {
		f.r = bufio.NewReaderSize(f.File, size)
		f.w = bufio.NewWriterSize(f.File, size)
	}

	return
}
