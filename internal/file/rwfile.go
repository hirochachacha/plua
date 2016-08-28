package file

import (
	"bufio"
	"bytes"
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
	f     *os.File
	wmode wmode
	state state // previous action type
	r     *bufio.Reader
	w     *bufio.Writer
}

func (f *file) Close() error {
	return f.f.Close()
}

func (f *file) Write(p []byte) (nn int, err error) {
	if f.state == read {
		_, err = f.Seek(-int64(f.r.Buffered()), 1)
		if err != nil {
			return
		}

		f.r.Reset(f.f)
	}

	switch f.wmode {
	case wno:
		nn, err = f.f.Write(p)
		if err != nil {
			return
		}
	case wfull:
		nn, err = f.w.Write(p)
		if err != nil {
			return
		}
	case wline:
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

		f.r.Reset(f.f)
	}

	switch f.wmode {
	case wno:
		nn, err = f.f.WriteString(s)
		if err != nil {
			return
		}
	case wfull:
		nn, err = f.w.WriteString(s)
		if err != nil {
			return
		}
	case wline:
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

		f.r.Reset(f.f)
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
		n, err = f.f.Seek(offset, 0)
	case 1:
		if f.state == read {
			n, err = f.f.Seek(offset-int64(f.r.Buffered()), 1)
			f.r.Reset(f.f)
		} else {
			n, err = f.f.Seek(offset, 1)
		}
	case 2:
		n, err = f.f.Seek(offset, 2)
	}

	if err != nil {
		return
	}

	f.state = seek

	return
}

func (f *file) SetVBuf(wmode wmode, size int) (err error) {
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

	f.wmode = wmode
	f.state = seek

	if size > 0 {
		f.r = bufio.NewReaderSize(f.f, size)
		f.w = bufio.NewWriterSize(f.f, size)
	}

	return
}
