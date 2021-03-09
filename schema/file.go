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

const maxMemSize = 1 << 24 // 16MB

type LineSource interface {
	// WriteLine writes a line. Given slice of bytes should not end with '\n'
	WriteLine([]byte) error

	Rand(r *rand.Rand) ([]byte, error)
}

type LineSourceFlusher interface {
	LineSource
	FlushTo(LineSource) error
}

type FallbackSource struct {
	LineSource
	fallback func() (*FallbackSource, error)
}

func (s FallbackSource) WriteLine(line []byte) error {
	err := s.LineSource.WriteLine(line)
	if err != nil {
		sf, ok := s.LineSource.(LineSourceFlusher)
		if !ok || s.fallback == nil {
			return err
		}
		s.LineSource, err = s.fallback()
		if err != nil {
			return fmt.Errorf("fallback: %w", err)
		}
		if err := sf.FlushTo(s.LineSource); err != nil {
			return fmt.Errorf("flush to fallback: %w", err)
		}
		if cl, ok := sf.(io.Closer); ok {
			if err := cl.Close(); err != nil {
				return fmt.Errorf("close: %w", err)
			}
		}
		return s.WriteLine(line)
	}
	return err
}

type bufferedSource struct {
	capacity uint64
	size     uint64
	lines    [][]byte
}

func newBufferedLineSource(capacity uint64) *bufferedSource {
	return &bufferedSource{
		capacity: capacity,
		lines:    make([][]byte, 0), // zero because we do not know the length of each string
	}
}

func (s *bufferedSource) Close() error {
	s.lines = nil
	s.size = 0
	return nil
}

func (f *bufferedSource) Rand(r *rand.Rand) ([]byte, error) {
	if len(f.lines) == 0 {
		return []byte{}, nil
	}
	return f.lines[r.Intn(len(f.lines))], nil
}

var ErrBufferFull = errors.New("buffer is full")

func (f *bufferedSource) WriteLine(line []byte) error {
	if f.size+uint64(len(line)) > f.capacity {
		return ErrBufferFull
	}
	f.lines = append(f.lines, line)
	f.size += uint64(len(line) - 1)
	return nil
}

func (f *bufferedSource) FlushTo(src LineSource) error {
	for len(f.lines) > 0 {
		if err := src.WriteLine(f.lines[0]); err != nil {
			return err
		}
		f.lines = f.lines[1:]
	}
	f.lines = nil // delete slice
	f.size = 0
	return nil
}

// index is a slice of indexes of ENDs of strings
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

func (f *indexedReaderAt) Rand(r *rand.Rand) ([]byte, error) {
	toInd := r.Intn(len(f.index))
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
	_, err := f.w.Write(append(line, '\n'))
	if err != nil {
		return err
	}
	return f.index.WriteLine(line)
}

func ScanFile(name string) (LineSource, error) {
	from, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	s, keepOpen, err := makeSource(from)
	if err != nil {
		return nil, err
	}
	if !keepOpen {
		defer from.Close()
	}

	sc := bufio.NewScanner(from)
	for sc.Scan() {

		if err := s.WriteLine([]byte(sc.Text())); err != nil {
			return nil, err
		}
	}
	return s, sc.Err()
}

func makeSource(f *os.File) (s LineSource, keepOpen bool, err error) {
	info, err := f.Stat()
	if err != nil {
		return nil, false, err
	}

	if info.Mode().IsRegular() {
		if info.Size() <= maxMemSize {
			return newBufferedLineSource(maxMemSize), false, nil
		}
		return newIndexedReaderAt(f, uint64(info.Size())), true, nil
	}
	bufs := newBufferedLineSource(maxMemSize)
	return FallbackSource{
		LineSource: bufs,
		fallback: func() (*FallbackSource, error) {
			tmp, err := ioutil.TempFile("", path.Base(f.Name()))
			if err != nil {
				return nil, fmt.Errorf("unable to create temporary file: %w", err)
			}
			return &FallbackSource{
				LineSource: newIndexedCopyReaderAt(tmp, uint64(len(bufs.lines))),
			}, nil
		},
	}, false, nil
}
