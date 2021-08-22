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
represent a stream of T as a pair of some value T and a function that produces
the next value of the stream T, and that function may returns a pointer, which
will be `nil` to represent the end of the stream (the function itself should
never be `nil` in this case). Stream values are pointers so the empty stream
can be represented as `nil`. Go doesn't directly support pairs, so we're going
to use a struct here:

```go
type Stream[T any] struct {
	Value T
	Next  func() *Stream[T]
}
```

It looks like a linked-list, but since we want to be able to represent a
potentially infinite linked list, instead of making `Next` a pointer to Stream,
we make it a function that returns the pointer, and lazily execute that
function as needed.

Alternatively, since Go supports returning multiple values from a function, we
could represent a stream as a function that returns the value and the function
to calculate the next value:

```go
type Stream[T any] func() (T, func() Stream[T])
```

Here, an empty stream is also represented as nil (no explicit pointers needed
though). Let's go crazy on the thought experiment and use only the function
based representation, instead of the linked list look-a-like.

## Creating streams

Let's do something probably non-sense: given a slice of some type `T`, how
would we generate a stream? Let's call that function `FromSlice`.

Here's what the code looks like:

```go
func FromSlice[T any](items []T) Stream[T] {
	return fromSlice(items, 0)
}

func fromSlice[T any](items []T, index int) Stream[T] {
	if index >= len(items) {
		return nil
	}
	return Stream[T](func() (T, func() Stream[T]) {
		return items[index], func() Stream[T] {
			return fromSlice(items, index+1)
		}
	})
}
```

To keep the context around of where we are in the slice as the stream is
consumed, we introduce a helper function that takes the index that we want to
consume in the slice, and `FromSlice` invokes that helper function starting at
index 0. The stream becomes nil as soon as the index grows too large,
indicating that we finished iterating over the slice.

### Infinite streams

Using a slice is no fun, we probably want to be able to implement potentially
infinite streams. For example, a stream of natural numbers could be represented
as follows:

```go
func nat(start int) Stream[int] {
	return Stream[int](func() (int, func() Stream[int]) {
		return start, func() Stream[int] {
			return nat(start + 1)
		}
	})
}
```

Here is where the limitation with Go's type inference starts to show: since it
can't infer return types, we have to manually specify `Stream[int]` in multiple
places.

Note how we never return

## Working with streams
