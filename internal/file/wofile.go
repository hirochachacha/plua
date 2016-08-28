package file

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

type wofile struct {
	f     *os.File
	wmode wmode
	w     *bufio.Writer
}

func (wo *wofile) Close() error {
	return wo.f.Close()
}

func (wo *wofile) Write(p []byte) (nn int, err error) {
	switch wo.wmode {
	case wno:
		nn, err = wo.f.Write(p)
	case wfull:
		nn, err = wo.w.Write(p)
	case wline:
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
	switch wo.wmode {
	case wno:
		nn, err = wo.f.WriteString(s)
	case wfull:
		nn, err = wo.w.WriteString(s)
	case wline:
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
	return ErrInvalid
}

func (wo *wofile) ReadByte() (c byte, err error) {
	return 0, ErrInvalid
}

func (wo *wofile) Read(p []byte) (n int, err error) {
	return 0, ErrInvalid
}

func (wo *wofile) ReadInt() (i int64, err error) {
	return 0, ErrInvalid
}

func (wo *wofile) ReadFloat() (f float64, err error) {
	return 0, ErrInvalid
}

func (wo *wofile) ReadSlice(delim byte) (line []byte, err error) {
	return nil, ErrInvalid
}

func (wo *wofile) Seek(offset int64, whence int) (n int64, err error) {
	err = wo.w.Flush()
	if err != nil {
		return
	}

	n, err = wo.f.Seek(offset, whence)

	return
}

func (wo *wofile) SetVBuf(wmode wmode, size int) (err error) {
	err = wo.w.Flush()
	if err != nil {
		return
	}

	wo.wmode = wmode

	if size > 0 {
		wo.w = bufio.NewWriterSize(wo.f, size)
	}

	return
}
