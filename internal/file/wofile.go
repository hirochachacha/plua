package file

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
)

type wofile struct {
	*os.File
	mode   int
	w      *bufio.Writer
	closed bool
	std    bool
}

func (wo *wofile) IsClosed() bool {
	return wo.closed
}

func (wo *wofile) Close() error {
	if wo.std {
		return errors.New("cannot close standard file")
	}

	err := wo.File.Close()

	wo.closed = true

	return err
}

func (wo *wofile) Write(p []byte) (nn int, err error) {
	switch wo.mode {
	case IONBF:
		nn, err = wo.File.Write(p)
	case IOFBF:
		nn, err = wo.w.Write(p)
	case IOLBF:
		i := bytes.LastIndex(p, []byte{'\n'})
		if i == -1 {
			nn, err = wo.w.Write(p)
		} else {
			nn, err = wo.w.Write(p[:i])
			if err != nil {
				return
			}
			err = wo.w.Flush()
			if err != nil {
				return
			}

			nn, err = wo.w.Write(p[i+1:])
		}
	}

	return
}

func (wo *wofile) WriteString(s string) (nn int, err error) {
	switch wo.mode {
	case IONBF:
		nn, err = wo.File.WriteString(s)
	case IOFBF:
		nn, err = wo.w.WriteString(s)
	case IOLBF:
		i := strings.LastIndex(s, "\n")
		if i == -1 {
			nn, err = wo.w.WriteString(s)
		} else {
			nn, err = wo.w.WriteString(s[:i])
			if err != nil {
				return
			}
			err = wo.w.Flush()
			if err != nil {
				return
			}

			nn, err = wo.w.WriteString(s[i+1:])
		}
	}

	return
}

func (wo *wofile) Flush() error {
	return wo.w.Flush()
}

func (wo *wofile) UnreadByte() error {
	return os.ErrInvalid
}

func (wo *wofile) ReadByte() (c byte, err error) {
	_, err = wo.Read(nil)

	return 0, err
}

func (wo *wofile) ReadInt() (i int64, err error) {
	_, err = wo.Read(nil)

	return 0, err
}

func (wo *wofile) ReadFloat() (f float64, err error) {
	_, err = wo.Read(nil)

	return 0, err
}

func (wo *wofile) ReadSlice(delim byte) (line []byte, err error) {
	_, err = wo.Read(nil)

	return nil, err
}

func (wo *wofile) Seek(offset int64, whence int) (n int64, err error) {
	err = wo.w.Flush()
	if err != nil {
		return
	}

	n, err = wo.File.Seek(offset, whence)

	return
}

func (wo *wofile) Setvbuf(mode int, size int) (err error) {
	err = wo.w.Flush()
	if err != nil {
		return
	}

	wo.mode = mode

	if size > 0 {
		wo.w = bufio.NewWriterSize(wo.File, size)
	}

	return
}
