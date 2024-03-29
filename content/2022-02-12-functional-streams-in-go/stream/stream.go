package stream

import (
	"bufio"
	"io"
	"strings"
)

type Stream[T any] struct {
	Value T
	Next  func() *Stream[T]
}

func Iter[T any](stream *Stream[T], f func(T)) {
	for ; stream != nil; stream = stream.Next() {
		f(stream.Value)
	}
}

func Fold[T, U any](stream *Stream[T], init U, f func(U, T) U) U {
	acc := init
	for ; stream != nil; stream = stream.Next() {
		acc = f(acc, stream.Value)
	}
	return acc
}

func Take[T any](stream *Stream[T], n uint) *Stream[T] {
	return TakeWhile(stream, func(T) bool {
		if n == 0 {
			return false
		}
		n--
		return true
	})
}

func Map[T, U any](stream *Stream[T], f func(T) U) *Stream[U] {
	if stream == nil {
		return nil
	}
	return &Stream[U]{
		Value: f(stream.Value),
		Next: func() *Stream[U] {
			return Map(stream.Next(), f)
		},
	}
}

func Append[T any](stream1 *Stream[T], stream2 *Stream[T]) *Stream[T] {
	if stream1 == nil {
		return stream2
	}
	return &Stream[T]{
		Value: stream1.Value,
		Next: func() *Stream[T] {
			return Append(stream1.Next(), stream2)
		},
	}
}

func FlatMap[T, U any](stream *Stream[T], f func(T) *Stream[U]) *Stream[U] {
	if stream == nil {
		return nil
	}
	return Append(f(stream.Value), FlatMap(stream.Next(), f))
}

func Filter[T any](stream *Stream[T], f func(T) bool) *Stream[T] {
	for ; stream != nil; stream = stream.Next() {
		if f(stream.Value) {
			return &Stream[T]{
				Value: stream.Value,
				Next: func() *Stream[T] {
					return Filter(stream.Next(), f)
				},
			}
		}
	}
	return stream
}

func TakeWhile[T any](stream *Stream[T], f func(T) bool) *Stream[T] {
	if stream == nil {
		return nil
	}
	if f(stream.Value) {
		return &Stream[T]{
			Value: stream.Value,
			Next: func() *Stream[T] {
				return TakeWhile(stream.Next(), f)
			},
		}
	}
	return nil
}

func TakeUntil[T any](stream *Stream[T], f func(T) bool) *Stream[T] {
	return TakeWhile(stream, func(v T) bool { return !f(v) })
}

func ToSlice[T any](stream *Stream[T]) []T {
	return Fold(stream, []T{}, func(acc []T, elm T) []T {
		return append(acc, elm)
	})
}

func FromSlice[T any](items []T) *Stream[T] {
	if len(items) == 0 {
		return nil
	}
	return &Stream[T]{
		Value: items[0],
		Next: func() *Stream[T] {
			return FromSlice(items[1:])
		},
	}
}

func FromReader(r io.Reader) *Stream[string] {
	return fromReader(bufio.NewReader(r))
}

func fromReader(r *bufio.Reader) *Stream[string] {
	line, isPrefix, err := r.ReadLine()
	if err != nil {
		return nil
	}
	parts := []string{string(line)}
	for isPrefix {
		line, isPrefix, err = r.ReadLine()
		if err != nil {
			return nil
		}
		parts = append(parts, string(line))
	}
	return &Stream[string]{
		Value: strings.Join(parts, ""),
		Next: func() *Stream[string] {
			return fromReader(r)
		},
	}
}

func Empty[T any]() *Stream[T] {
	return nil
}

func Singleton[T any](element T) *Stream[T] {
	return &Stream[T]{
		Value: element,
		Next:  Empty[T],
	}
}

func Repeat[T any](element T) *Stream[T] {
	return &Stream[T]{
		Value: element,
		Next: func() *Stream[T] {
			return Repeat(element)
		},
	}
}

func nat(start int) *Stream[int] {
	return &Stream[int]{
		Value: start,
		Next: func() *Stream[int] {
			return nat(start + 1)
		},
	}
}

func fromStrings(s *Stream[string]) *Stream[rune] {
	return FlatMap(s, func(v string) *Stream[rune] {
		return FromSlice([]rune(v))
	})
}
