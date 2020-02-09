package schema

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
)

type File struct {
	Path    string
	Scanned bool
	isTmp   bool
	file    *os.File
	delims  []int64
}

func (f *File) Set(name string) error {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return err
	}
	f.Path = name
	return nil
}

func (f *File) Type() string {
	return "string"
}

func (f *File) String() string {
	return f.Path
}

func (f *File) Scan() error {
	ff, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	stat, err := ff.Stat() // TODO: use from Set()
	if err != nil {
		return err
	}
	f.Scanned = true

	if f.delims == nil {
		f.delims = make([]int64, 0)
	}

	var cp func([]byte) error

	if stat.Mode().IsRegular() {
		f.file = ff
	} else {
		tmp, err := ioutil.TempFile("", path.Base(ff.Name()))
		if err != nil {
			return fmt.Errorf("unable to create temporary file: %w", err)
		}
		f.file = tmp
		f.isTmp = true
		w := bufio.NewWriterSize(tmp, 1024)
		defer w.Flush()
		cp = func(b []byte) error {
			nw, err := w.Write(b)
			if err != nil {
				return err
			}
			if nw != len(b) {
				return io.ErrShortWrite
			}
			return nil
		}

	}

	r := bufio.NewReaderSize(ff, 1024)
	for {
		bb, err := r.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if cp != nil {
			if err := cp(bb); err != nil {
				return err
			}
		}
		var next int64

		if len(f.delims) == 0 {
			next = int64(len(bb))
		} else {
			next = f.delims[len(f.delims)-1] + int64(len(bb))
		}
		if err == io.EOF {
			f.delims = append(f.delims, next+1)
			break
		}
		f.delims = append(f.delims, next)
	}
	return nil
}

func ScanFile(filePath string) (*File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	ff := &File{
		delims: make([]int64, 0),
	}

	var cp func([]byte) error

	if stat.Mode().IsRegular() {
		ff.file = f
	} else {
		tmp, err := ioutil.TempFile("", path.Base(f.Name()))
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary file: %w", err)
		}
		ff.file = tmp
		ff.isTmp = true
		w := bufio.NewWriterSize(tmp, 1024)
		defer w.Flush()
		cp = func(b []byte) error {
			nw, err := w.Write(b)
			if err != nil {
				return err
			}
			if nw != len(b) {
				return io.ErrShortWrite
			}
			return nil
		}

	}

	r := bufio.NewReaderSize(f, 1024)
	for {
		bb, err := r.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if cp != nil {
			if err := cp(bb); err != nil {
				return nil, err
			}
		}
		var next int64

		if len(ff.delims) == 0 {
			next = int64(len(bb))
		} else {
			next = ff.delims[len(ff.delims)-1] + int64(len(bb))
		}
		if err == io.EOF {
			ff.delims = append(ff.delims, next+1)
			break
		}
		ff.delims = append(ff.delims, next)
	}
	return ff, err
}

func (f *File) addDelim(numRead int) {
	if numRead == 0 {
		return
	}
}

func (f *File) Rand() (string, error) {
	toInd := rand.Intn(len(f.delims))
	var from int64
	if toInd == 0 {
		from = 0
	} else {
		from = f.delims[toInd-1]
	}
	buff := make([]byte, f.delims[toInd]-from-1)
	_, err := f.file.ReadAt(buff, from)
	return string(buff), err
}

func (f *File) Close() error {
	if err := f.file.Close(); err != nil {
		return err
	}
	if f.isTmp {
		if err := os.Remove(f.file.Name()); err != nil {
			return fmt.Errorf("failed to remove temporary file: %w", err)
		}
	}
	return nil
}
