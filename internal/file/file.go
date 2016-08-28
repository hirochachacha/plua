package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hirochachacha/blua/internal/strconv"
)

type OptionError struct {
	Opt string
}

func (opt *OptionError) Error() string {
	return ""
}

var (
	ErrInvalid = errors.New("invalid operation")

	ErrClosedFile = errors.New("attemp to use a closed file")
	ErrStdFile    = errors.New("cannot close standard file")
)

var (
	Stdin = &File{
		f:   newReadOnlyFileSize(os.Stdin, 0),
		std: true,
	}

	Stdout = &File{
		f:   newWriteOnlyFileMode(os.Stdout, wno),
		std: true,
	}

	Stderr = &File{
		f:   newWriteOnlyFileMode(os.Stderr, wno),
		std: true,
	}
)

type File struct {
	f   iFile
	std bool
}

func OpenFile(fname, mode string) (file *File, err error) {
	var m int

	switch mode {
	case "r":
		m = os.O_RDONLY

		f, err := os.OpenFile(fname, m, 0)
		if err != nil {
			return nil, err
		}
		file = &File{f: newReadOnlyFile(f)}
	case "w":
		m = os.O_WRONLY | os.O_TRUNC | os.O_CREATE

		f, err := os.OpenFile(fname, m, 0644)
		if err != nil {
			return nil, err
		}
		file = &File{f: newWriteOnlyFile(f)}
	case "a":
		m = os.O_WRONLY | os.O_APPEND | os.O_CREATE

		f, err := os.OpenFile(fname, m, 0644)
		if err != nil {
			return nil, err
		}
		file = &File{f: newWriteOnlyFile(f)}
	case "r+":
		m = os.O_RDWR

		f, err := os.OpenFile(fname, m, 0)
		if err != nil {
			return nil, err
		}
		file = &File{f: newFile(f)}
	case "w+":
		m = os.O_RDWR | os.O_TRUNC | os.O_CREATE

		f, err := os.OpenFile(fname, m, 0644)
		if err != nil {
			return nil, err
		}
		file = &File{f: newFile(f)}
	case "a+":
		m = os.O_RDWR | os.O_APPEND | os.O_CREATE

		f, err := os.OpenFile(fname, m, 0644)
		if err != nil {
			return nil, err
		}
		file = &File{f: newFile(f)}
	default:
		return nil, &OptionError{mode}
	}

	return
}

func TempFile(dir, prefix string) (*File, error) {
	tmp, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}

	return &File{f: newFile(tmp)}, nil
}

func Popen(prog, mode string) (*File, error) {
	args := strings.Fields(prog)

	cmd := exec.Command(args[0], args[1:]...)

	switch mode {
	case "r":
		f, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		file, ok := f.(*os.File)
		if !ok {
			return nil, errors.New("unexpected behavior")
		}

		return &File{f: newReadOnlyFile(file)}, nil
	case "w":
		f, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		file, ok := f.(*os.File)
		if !ok {
			return nil, errors.New("unexpected behavior")
		}

		return &File{f: newWriteOnlyFile(file)}, nil
	default:
		return nil, &OptionError{mode}
	}

	panic("unreachable")

	return nil, nil
}

func (f *File) Type() string {
	if f.f == nil {
		return "file (closed)"
	}

	return "file"
}

func (f *File) String() string {
	if f.f == nil {
		return "file (closed)"
	}

	return fmt.Sprintf("file (%p)", f)
}

func (f *File) Close() error {
	if f.f == nil {
		return ErrClosedFile
	}
	if f.std {
		return ErrStdFile
	}

	err := f.f.Close()

	f.f = nil

	return err
}

func (f *File) Flush() error {
	if f.f == nil {
		return ErrClosedFile
	}

	return f.f.Flush()
}

func (f *File) ReadFloat() (f64 float64, err error) {
	if f.f == nil {
		return 0, ErrClosedFile
	}

	return strconv.ScanFloat(f.f)
}

func (f *File) ReadInteger() (i64 int64, err error) {
	if f.f == nil {
		return 0, ErrClosedFile
	}

	return strconv.ScanInt(f.f)
}

func (f *File) ReadAll() (s string, err error) {
	if f.f == nil {
		return "", ErrClosedFile
	}

	bs, err := ioutil.ReadAll(f.f)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (f *File) ReadStripedLine() (s string, err error) {
	if f.f == nil {
		return "", ErrClosedFile
	}

	line, err := f.f.ReadSlice('\n')
	if err != nil {
		return "", err
	}

	return string(line[:len(line)-1]), nil
}

func (f *File) ReadLine() (s string, err error) {
	if f.f == nil {
		return "", ErrClosedFile
	}

	line, err := f.f.ReadSlice('\n')
	if err != nil {
		return "", err
	}

	return string(line), nil
}

func (f *File) Read(i int64) (s string, err error) {
	if f.f == nil {
		return "", ErrClosedFile
	}

	if f.f == nil {
		return "", ErrClosedFile
	}

	bs := make([]byte, i)

	n, err := f.f.Read(bs)
	if err != nil {
		return "", err
	}

	return string(bs[:n]), nil
}

func (f *File) Seek(whence string, offset int64) (ret int64, err error) {
	if f.f == nil {
		return 0, ErrClosedFile
	}

	switch whence {
	case "set":
		return f.f.Seek(offset, 0)
	case "cur":
		return f.f.Seek(offset, 1)
	case "end":
		return f.f.Seek(offset, 2)
	default:
		return 0, &OptionError{whence}
	}

	panic("unreachable")

	return
}

func (f *File) SetVBuf(wmode string, size int) error {
	if f.f == nil {
		return ErrClosedFile
	}

	switch wmode {
	case "no":
		f.f.SetVBuf(wno, size)
	case "line":
		f.f.SetVBuf(wline, size)
	case "full":
		f.f.SetVBuf(wfull, size)
	default:
		return &OptionError{wmode}
	}

	panic("unreachable")

	return nil
}

func (f *File) Write(s string) (n int, err error) {
	if f.f == nil {
		return 0, ErrClosedFile
	}

	return f.f.WriteString(s)
}
