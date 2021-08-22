package main

type Stream[T any] struct {
	Value T
	Next  func() *Stream[T]
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

func main() {
}
