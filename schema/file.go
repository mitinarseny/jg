package schema

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
)

type LineSource interface {
	// WriteLine accepts slice of bytes ending with \n
	WriteLine([]byte) error
	Rand() (string, error)
}

type BufferedLineSource interface {
	LineSource
	FlushTo(LineSource) error
}

var ErrBufferFull = errors.New("buffer is full")

type bufferedLineSource struct {
	capacity int64
	size     int64
	lines    []string
}

func newBufferedLineSource(capacity int64) *bufferedLineSource {
	return &bufferedLineSource{
		capacity: capacity,
		lines:    make([]string, 0),
	}
}

func (f *bufferedLineSource) Rand() (string, error) {
	if len(f.lines) == 0 {
		return "", nil
	}
	return f.lines[rand.Intn(len(f.lines))], nil
}

func (f *bufferedLineSource) WriteLine(line []byte) error {
	if f.size+int64(len(line)) > f.capacity {
		return ErrBufferFull
	}
	line = line[:len(line)-1] // trim newline
	f.lines = append(f.lines, string(line))
	f.size += int64(len(line))
	return nil
}

func (f *bufferedLineSource) FlushTo(src LineSource) error {
	for _, line := range f.lines {
		if err := src.WriteLine([]byte(line)); err != nil {
			return err
		}
		// TODO: reduce memory here
	}
	// TODO: f.size = 0
	return nil
}

type indexer []int64

func newIndexer(size int) indexer {
	return make(indexer, 0, size)
}

func (f *indexer) WriteLine(line []byte) error {
	if len(*f) == 0 {
		*f = append(*f, int64(len(line)))
	}
	*f = append(*f, (*f)[len(*f)-1]+int64(len(line)))
	return nil
}

type indexedFile struct {
	indexer
	io.ReaderAt
}

func newIndexedFile(r io.ReaderAt, size int) *indexedFile {
	return &indexedFile{
		indexer:  newIndexer(size),
		ReaderAt: r,
	}
}

func (f *indexedFile) Rand() (string, error) {
	toInd := rand.Intn(len(f.indexer))
	var from int64
	if toInd == 0 {
		from = 0
	} else {
		from = (f.indexer)[toInd-1]
	}
	buff := make([]byte, f.indexer[toInd]-from-1)
	_, err := f.ReadAt(buff, from)
	return string(buff), err
}

type indexedCopyFile struct {
	*indexedFile
	io.Writer
}

func newIndexedCopyFile(file interface {
	io.Writer
	io.ReaderAt
}, size int) *indexedCopyFile {
	return &indexedCopyFile{
		indexedFile: newIndexedFile(file, size),
		Writer:      file,
	}
}

func (f *indexedCopyFile) WriteLine(line []byte) error {
	n, err := f.Write(line)
	if err != nil {
		return err
	}
	if n != len(line) {
		return io.ErrShortWrite
	}
	return f.indexer.WriteLine(line)
}

type File struct {
	from *os.File
	LineSource
	Path      string
	Scanned   bool
	isRegular bool
}

func (f *File) Set(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	f.isRegular = info.Mode().IsRegular()
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
	f.Scanned = true

	// TODO: do not create buffer if file size is greater than 1024
	b := newBufferedLineSource(1024) // TODO: parametrize size
	f.LineSource = b

	r := bufio.NewReaderSize(ff, 1024)
	for {
		bb, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				return f.Register(append(bb, '\n'))
			}
			return err
		}
		if err := f.Register(bb); err != nil {
			return err
		}
	}
}

func (f *File) Register(line []byte) error {
	if err := f.LineSource.WriteLine(line); err != nil {
		if err != ErrBufferFull {
			return err
		}
		buff, ok := f.LineSource.(*bufferedLineSource)
		if !ok {
			return err
		}
		if f.isRegular {
			f.LineSource = newIndexedFile(f.from, len(buff.lines))
		} else {
			tmp, err := ioutil.TempFile("", path.Base(f.from.Name())) // TODO: basename pattern
			if err != nil {
				return fmt.Errorf("unable to create temporary file: %w", err)
			}
			f.LineSource = newIndexedCopyFile(tmp, len(buff.lines))
		}

		if err := buff.FlushTo(f.LineSource); err != nil {
			return err
		}

	}
	return nil
}

func (f *File) Close() error {
	if err := f.from.Close(); err != nil {
		return err
	}
	if !f.isRegular { // pipe can not be deleted
		if err := os.Remove(f.from.Name()); err != nil {
			return fmt.Errorf("failed to remove temporary file: %w", err)
		}
	}
	return nil
}
