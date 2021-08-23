+++
title = "Implementing functional streams with generics in Go (for fun)"

draft = true

[taxonomies]
tags = ["golang", "go", "generics"]
+++

> **Note:** this is an experiment with an upcoming change in Go's type system.
> If you need something like what's described in the post in the real world,
> use channels.

Go 1.18 will likely support generics, and I decided I would give it a shot.
I've been playing with the idea of [representing streams with
channels](https://github.com/fsouza/channels) and higher order functions that
operate on those channels, allowing processes to execute on streams of data
somewhat-lazily. This is 100% a toy project and a production-ready version of
"lazy/generic streams" will likely come in some future release of
[RxGo](https://github.com/ReactiveX/RxGo).

Still, while working on a solution to get this working with channels, I figured
we could also try to get it working in a more traditional representation, at
least in functional languages: streams can be represented by a pair - a value
and a function to generate the next value.

I started with a question: with generics, is the Go type-system expressive
enough to represent such streams. And how bad would it look without Ocaml's or
Haskell's type inference capabilities? Turns out we _can_ represent streams,
and the limited type-inference makes it a bit annoying, but it isn't too bad!

## The basic type definition

First, let's start with the type definition. Like mentioned above, we want to
represent a stream of T as a nullable pair of some value T and a function that
produces the next value of the stream T. A `nil` stream represents an empty
stream.

Go doesn't directly support pairs, so we're going to use a struct here:

```go
type Stream[T any] struct {
	Value T
	Next  func() *Stream[T]
}
```

And we'll represent streams as pointers to that struct, which takes care of the
`nullable` part.

It looks like a linked-list, but since we want to be able to represent a
potentially infinite linked list, instead of making `Next` a pointer to Stream,
we make it a function that returns the pointer, and lazily execute that
function as needed.

## Creating streams

First, let's start with some very basic streams: an empty stream, and a
singleton stream (that contains a single element). We'll use helper functions
to implement those:

```go
func Empty[T any]() *Stream[T] {
	return nil
}

func Singleton[T any](element T) *Stream[T] {
	return &Stream[T]{
		Value: element,
		Next:  Empty[T],
	}
}
```

A more interesting helper function is a function that takes a slice and returns
a stream with the elements of that slice:

```go
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
```

To keep the context around of where we are in the slice as the stream is
consumed, we introduce a helper function that takes the index that we want to
consume in the slice, and `FromSlice` invokes that helper function starting at
index 0. The stream becomes nil as soon as the index grows too large,
indicating that we finished iterating over the slice.

### Infinite streams

Using a slice is no fun, we probably want to be able to implement streams that
are potentially infinite. For example, a stream of natural numbers could be
represented as follows:

```go
func nat(start int) *Stream[int] {
	return &Stream[int]{
		Value: start,
		Next: func() *Stream[int] {
			return nat(start + 1)
		},
	}
}
```

Here is where the limitation with Go's type inference starts to show: since it
can't infer return types, we have to manually specify `Stream[int]` in multiple
places.

Note how we never return `nil` in the function above, indicating that this
stream doesn't really end.

## Manipulating streams

Now that we know how to create them, we need to understand how to manipulate
them to accomplish something useful. Two useful helper functions, useful for
debugging and what not, are `Iter` and `ToSlice`: `Iter` takes a stream and a
function, and iterates over the stream, invoking the function for each element,
while `ToSlice` converts a stream to a slice.

Here's `Iter`:

```go
func Iter[T any](stream *Stream[T], f func(T)) {
	for ; stream != nil; stream = stream.Next() {
		f(stream.Value)
	}
}
```

And here's `ToSlice`, built on top of `Iter`:

```go
func ToSlice[T any](stream *Stream[T]) []T {
	var result []T
	Iter(stream, func(value T) {
		result = append(result, value)
	})
	return result
}
```

(folks who are paying attention will probably suggest that we use something
like `Fold` instead of `Iter` to implement `ToSlice`, we'll get there).

And now that we have `Iter` and an infinite stream, we should try to use it
maybe? :)

```go
func main() {
	Iter(nat(0), func(v int) {
		fmt.Println(v)
	})
}
```

## Higher-order functions

TODO: describe the basic operations.

## Putting it all together in a real-ish example

TODO: write example using the higher-order functions, wrap up the post.
