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

type LineWriter interface {
	// WriteLine accepts slice of bytes ending with '\n'
	WriteLine([]byte) error
	Rand() (string, error)
}

type BufferedLineSource interface {
	LineWriter
	FlushTo(LineWriter) error
}

type bufferedLineSource struct {
	capacity uint64
	size     uint64
	lines    []string
}

func newBufferedLineSource(capacity uint64) *bufferedLineSource {
	return &bufferedLineSource{
		capacity: capacity,
		lines:    make([]string, 0), // zero because we do not know the length of each string
	}
}

func (f *bufferedLineSource) Rand() (string, error) {
	if len(f.lines) == 0 {
		return "", nil
	}
	return f.lines[rand.Intn(len(f.lines))], nil
}

func (f *bufferedLineSource) WriteLine(line []byte) error {
	l := len(line)
	if f.size+uint64(l) > f.capacity {
		return ErrBufferFull
	}
	line = line[:l-1] // trim newline
	f.lines = append(f.lines, string(line))
	f.size += uint64(l)
	return nil
}

func (f *bufferedLineSource) FlushTo(src LineWriter) error {
	for _, line := range f.lines {
		if err := src.WriteLine([]byte(line + "\n")); err != nil {
			return err
		}
		// TODO: reduce memory here
	}
	// TODO: f.size = 0
	return nil
}

var ErrBufferFull = errors.New("buffer is full")

type fallbackSource struct {
	Main       BufferedLineSource
	fallback   LineWriter
	FallbackFn func() (LineWriter, error)
}

func (s *fallbackSource) doFallback() error {
	var err error
	switch s.fallback, err = s.FallbackFn(); {
	case err != nil:
		return err
	case s.fallback == nil:
		panic("nil fallback")
	}

	if err := s.Main.FlushTo(s.fallback); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	if cl, ok := s.Main.(io.Closer); ok {
		if err := cl.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
	}

	s.Main = nil
	s.FallbackFn = nil
	return nil
}

func (s *fallbackSource) lineWriter() LineWriter {
	switch {
	case s.Main != nil:
		return s.Main
	case s.fallback != nil:
		return s.fallback
	default:
		panic("main and fallback are both nil")
	}
}

func (s *fallbackSource) WriteLine(line []byte) error {
	if s.fallback != nil {
		return s.fallback.WriteLine(line)
	}
	if err := s.Main.WriteLine(line); err == nil || s.FallbackFn == nil {
		return err
	}
	if err := s.doFallback(); err != nil {
		return fmt.Errorf("fallback: %w", err)
	}
	return s.WriteLine(line)
}

func (s *fallbackSource) Rand() (string, error) {
	return s.lineWriter().Rand()
}

func (s *fallbackSource) close(lw LineWriter) error {
	if cl, ok := lw.(io.Closer); ok {
		return cl.Close()
	}
	return nil
}

func (s *fallbackSource) Close() error {
	switch {
	case s.Main != nil:
		return s.close(s.Main)
	case s.fallback != nil:
		return s.close(s.fallback)
	}
	return nil
}

type index []int64

func newIndexer(size int) index {
	return make(index, 0, size)
}

func (f *index) WriteLine(line []byte) error {
	var prev int64
	if l := len(*f); l > 0 {
		prev = (*f)[l-1]
	}
	*f = append(*f, prev+int64(len(line)))
	return nil
}

type indexedReaderAt struct {
	index
	io.ReaderAt
}

func newIndexedReaderAt(r io.ReaderAt, size int) *indexedReaderAt {
	return &indexedReaderAt{
		index:    newIndexer(size),
		ReaderAt: r,
	}
}

func (f *indexedReaderAt) Rand() (string, error) {
	toInd := rand.Intn(len(f.index))
	var from int64
	if toInd > 0 {
		from = f.index[toInd-1]
	}
	buff := make([]byte, f.index[toInd]-from-1)
	_, err := f.ReadAt(buff, from)
	return string(buff), err
}

type indexedCopyReaderAt struct {
	*indexedReaderAt
	w io.Writer
}

func newIndexedCopyReaderAt(f interface {
	io.Writer
	io.ReaderAt
}, size int) *indexedCopyReaderAt {
	return &indexedCopyReaderAt{
		indexedReaderAt: newIndexedReaderAt(f, size),
		w:               f,
	}
}

func (f *indexedCopyReaderAt) WriteLine(line []byte) error {
	n, err := f.w.Write(line)
	if err != nil {
		return err
	}
	if n != len(line) {
		return io.ErrShortWrite
	}
	return f.index.WriteLine(line)
}

type File struct {
	from *os.File
	tmp  *os.File
	LineWriter
	Path     string
	scanned  bool
	seekable bool
}

func (f *File) Set(name string) (err error) {
	f.Path = name
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	f.seekable = info.Mode().IsRegular()
	return nil
}

func (f *File) Type() string {
	return "string"
}

func (f *File) String() string {
	return f.Path
}

func (f *File) Scanned() bool {
	return f.scanned
}

func (f *File) Scan() (err error) {
	f.from, err = os.Open(f.Path)
	if err != nil {
		return err
	}
	f.scanned = true

	// TODO: do not create buffer if file size is greater than 1024
	b := newBufferedLineSource(1) // TODO: parametrize size
	fbs := &fallbackSource{Main: b}

	if f.seekable {
		fbs.FallbackFn = func() (LineWriter, error) {
			return newIndexedReaderAt(f.from, len(b.lines)), nil
		}
	} else {
		fbs.FallbackFn = func() (LineWriter, error) {
			tmp, err := ioutil.TempFile("", path.Base(f.from.Name()))
			if err != nil {
				return nil, fmt.Errorf("unable to create temporary file: %w", err)
			}
			f.tmp = tmp
			return newIndexedCopyReaderAt(tmp, len(b.lines)), nil
		}
	}

	f.LineWriter = fbs

	r := bufio.NewReaderSize(f.from, 1024)
	for {
		bb, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				return f.WriteLine(append(bb, '\n'))
			}
			return err
		}
		if err := f.WriteLine(bb); err != nil {
			return err
		}
	}
}

func (f *File) Close() error {
	if err := f.from.Close(); err != nil {
		return err
	}
	if f.tmp != nil { // pipe can not be deleted
		if err := os.Remove(f.tmp.Name()); err != nil {
			return fmt.Errorf("failed to remove temporary file: %w", err)
		}
	}
	return nil
}
