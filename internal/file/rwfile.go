package file

import (
	"bufio"
	"bytes"
	"errors"
	"io"
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
	br     *bufio.Reader
	bw     *bufio.Writer
	off    int64
	mode   int
	state  state // previous action type
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

	if err := f.bw.Flush(); err != nil {
		return err
	}

	if err := f.File.Close(); err != nil {
		return err
	}

	f.closed = true

	return nil
}

func (f *file) Write(p []byte) (nn int, err error) {
	if f.state == read {
		_, err = f.Seek(f.off, io.SeekStart)
		if err != nil {
			return
		}

		f.br.Reset(f.File)
	}

	switch f.mode {
	case IONBF:
		nn, err = f.File.Write(p)
		f.off += int64(nn)
		if err != nil {
			return
		}
	case IOFBF:
		nn, err = f.bw.Write(p)
		f.off += int64(nn)
		if err != nil {
			return
		}
	case IOLBF:
		i := bytes.LastIndexByte(p, '\n')
		if i == -1 {
			nn, err = f.bw.Write(p)
			f.off += int64(nn)
			if err != nil {
				return
			}
		} else {
			nn, err = f.bw.Write(p[:i+1])
			f.off += int64(nn)
			if err != nil {
				return
			}
			err = f.bw.Flush()
			if err != nil {
				return
			}

			nn, err = f.bw.Write(p[i+1:])
			f.off += int64(nn)
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
		_, err = f.Seek(f.off, io.SeekStart)
		if err != nil {
			return
		}

		f.br.Reset(f.File)
	}

	switch f.mode {
	case IONBF:
		nn, err = f.File.WriteString(s)
		f.off += int64(nn)
		if err != nil {
			return
		}
	case IOFBF:
		nn, err = f.bw.WriteString(s)
		f.off += int64(nn)
		if err != nil {
			return
		}
	case IOLBF:
		i := strings.LastIndexByte(s, '\n')
		if i == -1 {
			nn, err = f.bw.WriteString(s)
			f.off += int64(nn)
			if err != nil {
				return
			}
		} else {
			nn, err = f.bw.WriteString(s[:i+1])
			f.off += int64(nn)
			if err != nil {
				return
			}
			err = f.bw.Flush()
			if err != nil {
				return
			}

			nn, err = f.bw.WriteString(s[i+1:])
			f.off += int64(nn)
			if err != nil {
				return
			}
		}
	}

	f.state = write

	return
}

func (f *file) Flush() error {
	return f.bw.Flush()
}

func (f *file) UnreadByte() (err error) {
	if f.state == write {
		err = f.bw.Flush()
		if err != nil {
			return
		}
	}

	err = f.br.UnreadByte()
	if err == nil {
		f.off--
	}

	f.state = read

	return
}

func (f *file) ReadByte() (c byte, err error) {
	if f.state == write {
		err = f.bw.Flush()
		if err != nil {
			return
		}
	}

	c, err = f.br.ReadByte()
	if err == nil {
		f.off++
	}

	f.state = read

	return
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.state == write {
		err = f.bw.Flush()
		if err != nil {
			return
		}
	}

	n, err = f.br.Read(p)
	f.off += int64(n)
	if err != nil {
		return
	}

	f.state = read

	return
}

func (f *file) ReadBytes(delim byte) (line []byte, err error) {
	if f.state == write {
		err = f.bw.Flush()
		if err != nil {
			return
		}

		f.br.Reset(f.File)
	}

	line, err = f.br.ReadBytes(delim)
	f.off += int64(len(line))
	if err != nil {
		return
	}

	f.state = read

	return
}

func (f *file) Seek(offset int64, whence int) (n int64, err error) {
	if f.state == write {
		if err := f.bw.Flush(); err != nil {
			return 0, err
		}
	}

	switch whence {
	case io.SeekStart:
		if f.off <= offset && offset <= f.off+int64(f.br.Buffered()) {
			f.br.Discard(int(offset - f.off))
			f.off = offset
		} else {
			f.off, err = f.File.Seek(offset, io.SeekStart)
			f.br.Reset(f.File)
		}
	case io.SeekCurrent:
		if 0 <= offset && offset <= int64(f.br.Buffered()) {
			f.br.Discard(int(offset))
			f.off += offset
		} else {
			f.off, err = f.File.Seek(f.off+offset, io.SeekStart)
			f.br.Reset(f.File)
		}
	case io.SeekEnd:
		f.off, err = f.File.Seek(offset, io.SeekEnd)
		f.br.Reset(f.File)
	}

	n = f.off

	if err != nil {
		return
	}

	f.state = seek

	return
}

func (f *file) Setvbuf(mode int, size int) (err error) {
	switch f.state {
	case read:
		_, err = f.Seek(-int64(f.br.Buffered()), 1)
		if err != nil {
			return err
		}
	case write:
		err = f.bw.Flush()
		if err != nil {
			return err
		}
	}

	f.mode = mode
	f.state = seek

	if size > 0 {
		f.br = bufio.NewReaderSize(f.File, size)
		f.bw = bufio.NewWriterSize(f.File, size)
	}

	return
}
