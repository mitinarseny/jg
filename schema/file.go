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
	Rand() ([]byte, error)
}

type BufferedLineSource interface {
	LineWriter
	FlushTo(LineWriter) error
}

type bufferedLineSource struct {
	capacity uint64
	size     uint64
	lines    [][]byte
}

func newBufferedLineSource(capacity uint64) *bufferedLineSource {
	return &bufferedLineSource{
		capacity: capacity,
		lines:    make([][]byte, 0), // zero because we do not know the length of each string
	}
}

func (f *bufferedLineSource) Rand() ([]byte, error) {
	if len(f.lines) == 0 {
		return []byte{}, nil
	}
	return f.lines[rand.Intn(len(f.lines))], nil
}

func (f *bufferedLineSource) WriteLine(line []byte) error {
	l := len(line)
	if f.size+uint64(l) > f.capacity {
		return ErrBufferFull
	}
	line = line[:l-1] // trim newline
	f.lines = append(f.lines, line)
	f.size += uint64(l)
	return nil
}

func (f *bufferedLineSource) FlushTo(src LineWriter) error {
	for _, line := range f.lines {
		if err := src.WriteLine(append(line, '\n')); err != nil {
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

func (s *fallbackSource) Rand() ([]byte, error) {
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

func newIndexer(size uint64) index {
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

func newIndexedReaderAt(r io.ReaderAt, size uint64) *indexedReaderAt {
	return &indexedReaderAt{
		index:    newIndexer(size),
		ReaderAt: r,
	}
}

func (f *indexedReaderAt) Rand() ([]byte, error) {
	toInd := rand.Intn(len(f.index))
	var from int64
	if toInd > 0 {
		from = f.index[toInd-1]
	}
	buff := make([]byte, f.index[toInd]-from-1)
	_, err := f.ReadAt(buff, from)
	return buff, err
}

type indexedCopyReaderAt struct {
	*indexedReaderAt
	w io.Writer
}

func newIndexedCopyReaderAt(f interface {
	io.Writer
	io.ReaderAt
}, size uint64) *indexedCopyReaderAt {
	return &indexedCopyReaderAt{
		indexedReaderAt: newIndexedReaderAt(f, size),
		w:               f,
	}
}

func (f *indexedCopyReaderAt) WriteLine(line []byte) error {
	_, err := f.w.Write(line)
	if err != nil {
		return err
	}
	return f.index.WriteLine(line)
}

type file struct {
	f *os.File
	isTmp bool
	LineWriter
}

func ScanFile(name string) (f *file, err error) {
	f = new(file)
	from, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := from.Stat()
	if err != nil {
		return nil, err
	}

	if info.Mode().IsRegular() && info.Size() > 1024 {
		f.f = from
		f.LineWriter = newIndexedReaderAt(f.f, uint64(info.Size()))
	} else {
		b := newBufferedLineSource(1024) // TODO: parametrize size?
		fbs := &fallbackSource{Main: b}
		if info.Mode().IsRegular() {
			fbs.FallbackFn = func() (LineWriter, error) {
				f.f = from
				return newIndexedReaderAt(f.f, uint64(len(b.lines))), nil
			}
		} else {
			fbs.FallbackFn = func() (LineWriter, error) {
				tmp, err := ioutil.TempFile("", path.Base(from.Name()))
				if err != nil {
					return nil, fmt.Errorf("unable to create temporary file: %w", err)
				}
				f.f = tmp
				f.isTmp = true
				return newIndexedCopyReaderAt(f.f, uint64(len(b.lines))), nil
			}
		}
		f.LineWriter = fbs
	}
	defer func() {
		if err != nil || from != f.f {
			_ = from.Close() // close from since we do not need it anymore
		}
	}()

	r := bufio.NewReaderSize(from, 1024) // TODO: parametrize size?
	for {
		bb, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				if err := f.WriteLine(append(bb, '\n')); err != nil {
					return nil, err
				}
				return f, nil
			}
			return nil, err
		}
		if err := f.WriteLine(bb); err != nil {
			return nil, err
		}
	}
}

func (f *file) Close() error {
	if f.f != nil {
		return f.f.Close()
	}
	if f.isTmp {
		return os.Remove(f.f.Name())
	}
	return nil
}
