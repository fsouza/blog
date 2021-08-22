+++
title = "Implementing functional streams with generics in Go for fun"

draft = true

[taxonomies]
tags = ["golang", "go", "generics"]
+++

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
Haskell's type inference capabilities?

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

Alternatively, since Go supports returning multiple values from a function, we
could represent a stream as a function that returns the value and the function
to calculate the next value:

```go
type Stream[T any] func() (T, func() Stream[T])
```

Here, an empty stream is also represented as nil (no explicit pointers needed
though). Let's explore both options in this article!

## Creating streams

## Working with streams
