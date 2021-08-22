package main

type Stream[T any] func() (T, func() Stream[T])

func Take[T any](stream Stream[T], n uint) Stream[T] {
	if n == 0 {
		return nil
	}
	return Stream[T](func() (T, func() Stream[T]) {
		v, tl := stream()
		return v, func() Stream[T] { return Take(tl(), n-1) }
	})
}

func Map[T, U any](stream Stream[T], f func(T) U) Stream[U] {
	if stream == nil {
		return nil
	}
	return Stream[U](func() (U, func() Stream[U]) {
		v, tl := stream()
		return f(v), func() Stream[U] { return Map(tl(), f) }
	})
}

func ToSlice[T any](stream Stream[T]) []T {
	var result []T
	for stream != nil {
		v, tl := stream()
		result = append(result, v)
		stream = tl()
	}
	return result
}

func main() {
}
