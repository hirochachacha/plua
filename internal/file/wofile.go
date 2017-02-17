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
	bw     *bufio.Writer
	mode   int
	closed bool
	std    bool
	probe  [1]byte // workaround for https://github.com/golang/go/issues/19122
}

func (wo *wofile) IsClosed() bool {
	return wo.closed
}

func (wo *wofile) Close() error {
	if wo.std {
		return errors.New("cannot close standard file")
	}

	if err := wo.bw.Flush(); err != nil {
		return err
	}

	if err := wo.File.Close(); err != nil {
		return err
	}

	wo.closed = true

	return nil
}

func (wo *wofile) Write(p []byte) (nn int, err error) {
	switch wo.mode {
	case IONBF:
		nn, err = wo.File.Write(p)
	case IOFBF:
		nn, err = wo.bw.Write(p)
	case IOLBF:
		i := bytes.LastIndexByte(p, '\n')
		if i == -1 {
			nn, err = wo.bw.Write(p)
		} else {
			nn, err = wo.bw.Write(p[:i+1])
			if err != nil {
				return
			}
			err = wo.bw.Flush()
			if err != nil {
				return
			}

			nn, err = wo.bw.Write(p[i+1:])
		}
	}

	return
}

func (wo *wofile) WriteString(s string) (nn int, err error) {
	switch wo.mode {
	case IONBF:
		nn, err = wo.File.WriteString(s)
	case IOFBF:
		nn, err = wo.bw.WriteString(s)
	case IOLBF:
		i := strings.LastIndexByte(s, '\n')
		if i == -1 {
			nn, err = wo.bw.WriteString(s)
		} else {
			nn, err = wo.bw.WriteString(s[:i+1])
			if err != nil {
				return
			}
			err = wo.bw.Flush()
			if err != nil {
				return
			}

			nn, err = wo.bw.WriteString(s[i+1:])
		}
	}

	return
}

func (wo *wofile) Flush() error {
	return wo.bw.Flush()
}

func (wo *wofile) UnreadByte() error {
	return os.ErrInvalid
}

func (wo *wofile) ReadByte() (c byte, err error) {
	_, err = wo.Read(wo.probe[:])

	return 0, err
}

func (wo *wofile) ReadBytes(delim byte) (line []byte, err error) {
	_, err = wo.Read(wo.probe[:])

	return nil, err
}

func (wo *wofile) Seek(offset int64, whence int) (n int64, err error) {
	if err := wo.bw.Flush(); err != nil {
		return 0, err
	}

	return wo.File.Seek(offset, whence)
}

func (wo *wofile) Setvbuf(mode int, size int) (err error) {
	err = wo.bw.Flush()
	if err != nil {
		return
	}

	wo.mode = mode

	if size > 0 {
		wo.bw = bufio.NewWriterSize(wo.File, size)
	}

	return
}
