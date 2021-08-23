package main

type Stream[T any] struct {
	Value T
	Next  func() *Stream[T]
}

func End[T any]() *Stream[T] {
	return nil
}

func Take[T any](stream *Stream[T], n uint) *Stream[T] {
	if n == 0 || stream == nil {
		return nil
	}
	return &Stream[T]{
		Value: stream.Value,
		Next: func() *Stream[T] {
			return Take(stream.Next(), n-1)
		},
	}
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

func ToSlice[T any](stream *Stream[T]) []T {
	var result []T
	for stream != nil {
		result = append(result, stream.Value)
		stream = stream.Next()
	}
	return result
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

func FromSlice[T any](items []T) *Stream[T] {
	return fromSlice(items, 0)
}

func fromSlice[T any](items []T, index int) *Stream[T] {
	if index >= len(items) {
		return nil
	}
	return &Stream[T]{
		Value: items[index],
		Next: func() *Stream[T] {
			return fromSlice(items, index+1)
		},
	}
}

func main() {
}
