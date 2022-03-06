
+++
title = "Implementing functional streams with generics in Go"

[taxonomies]
tags = ["golang", "go", "generics"]
+++

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [The basic type definition](#the-basic-type-definition)
- [Creating streams](#creating-streams)
  - [Infinite streams](#infinite-streams)
- [Manipulating streams](#manipulating-streams)
- [Higher-order functions](#higher-order-functions)
  - [Map](#map)
  - [Filter](#filter)
  - [Fold](#fold)
  - [FlatMap](#flatmap)
- [Slicing streams](#slicing-streams)
- [Putting it all together in a realistic example](#putting-it-all-together-in-a-realistic-example)
  - [Another more interesting example](#another-more-interesting-example)
- [Why not methods?](#why-not-methods)
- [Feedback](#feedback)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

> **Note 1:** this is an experiment with an upcoming change in Go's type system.
> If you need something like what's described in the post in the real world,
> use channels, or something like RxGo.
>
> **Note 2:** I originally started this post in August and almost abandoned it,
> but figured it's still a useful exploration of an important upcoming feature
> use channels, or something like RxGo.


Go 1.18 will support generics, and I decided I would give it a shot. I've been
playing with the idea of [representing streams with
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

One potentially

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

This is a very fancy way of making an infinite loop :)

## Higher-order functions

Higher-order functions make streams more interesting. There are many possible
high-order functions, but we'll explore some common names here: `Map`, `Filter`
and `Fold` (`Fold` may also be called `Reduce` in other contexts). Then we'll
do a fun one with `FlatMap`.

> **Note:** different from actual lists/arrays, with streams one can't
> implement `Map` or `Filter` using `Fold`, as `Fold` always consumes the
> stream.

The interesting thing with streams is that functions such as `Filter` and `Map`
don't do any computation unless needed: streams are lazy by nature, and actual
computations only happen when they're consumed.

### Map

`Map` takes a stream of type `T` and a function from `T` to `U` and returns a
stream of `U`. Here's the implementation of `Map`:

```go
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
```

How can we use `Map`? A simple/stupid example would be to double all numbers
from the `nat` stream to get a stream of even numbers:

```go
func main() {
	evens := Map(nat(0), func(v int) int {
		return v * 2
	})
	Iter(evens, func(v int) {
		fmt.Println(v)
	})
}
```

### Filter

`Filter` takes a stream of type `T` and a predicate function from `T` to `bool`
and returns a new stream of `T`, containing only the elements of the original
stream for which the given predicate returns `true`. Here's the implementation
of `filter`:

```go
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
```

Since a non-nil Stream is required to have a valid element, `Filter` isn't
totally lazy, as it has to consume the source stream until the predicate `p`
returns `true`. Here's how we can get a stream of even numbers using `Filter`
instead of `Map`:

```go
func main() {
	evens := Filter(nat(0), func(v int) bool {
		return v%2 == 0
	})
	Iter(evens, func(v int) {
		fmt.Println(v)
	})
}
```

### Fold

`Fold` can be used to eagerly fold the elements of a stream into some other
value. For example, if you have a finite stream of integers, you could use
`Fold` to find the largest, the smallest or the sum of all elements in the
stream.

Since `Fold` is eager, it can't really operate on infinite streams, as that
would loop forever. Let's look at the implementation of `Fold`:

```go
func Fold[T, U any](stream *Stream[T], init U, f func(U, T) U) U {
	acc := init
	for ; stream != nil; stream = stream.Next() {
		acc = f(acc, stream.Value)
	}
	return acc
}
```

And here's an implementation of `ToSlice` that uses `Fold` instead of `Iter`:

```go
func ToSlice[T any](stream *Stream[T]) []T {
	return Fold(stream, []T{}, func(acc []T, elm T) []T {
		return append(acc, elm)
	})
}
```

### FlatMap

`FlatMap` works like `Map`, but instead of taking a function from `T` to `U`,
it takes a function from `T` to `Stream[U]`. In order to implement `FlatMap`,
we'll first implement another useful function: `Append`, which takes two
streams `s1` and `s2` and returns a stream that will have all elements from
`s1`, then all elements from `s2`. Here's the code for both `Append` and
`FlatMap`:

```go
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
```

One simple-ish example is taking a stream of strings and turning that into a
stream of runes:

```go
var s *Stream[string] = ...
FlatMap(s, func(v string) *Stream[rune] {
	return FromSlice([]rune(v))
})
```

## Slicing streams

On top of filtering, mapping, appending and others, one may want to combine
multiple streams (using `Append` declared above), or get some elements of a
stream, or drop some items from a stream. For that, let's look at how we'd
implement some other helper functions:

- `Take(s Stream[T], n int) Stream[T]`: given stream `s`, returns a new stream
  with at most `n` elements.
- `TakeWhile(s Stream[T], p func(T) bool) Stream[T]`: given stream `s` and a
  predicate `p`, returns a new stream that will have elements from `s` as long
  as `p` returns `true`
- `TakeUntil(s Stream[T], p func(T) bool) Stream[T]`: like `TakeWhile`, but the
  output stream will have elements from `s` until `p` returns `true`.

Here's the source for `Take`, `TakeWhile` and `TakeUntil`:

```go
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
```

Notice how `TakeUntil` is simply implemented in terms of `TakeWhile`. As a
matter of fact, `Take` could also be implemented with `TakeWhile` and a
closure:

```go
func Take[T any](stream *Stream[T], n uint) *Stream[T] {
	return TakeWhile(stream, func(T) bool {
		if n == 0 {
			return false
		}
		n--
		return true
	})
}
```

## Putting it all together in a realistic example

Let's do a very simple REPL style application: it has a shell where we can
enter commands, and those commands will have some side-effect. Let's implement
a poor man [memcached](https://memcached.org/) that operates via stdin, and
supports three commands `set`, `get` and `del` (which will set a key-value
pair, get the value of a key and delete a key, respectively). We want to be
able to have two separate layers: one for parsing and another one for executing
commands, and we want to be able to send commands from stdin to the execution
layer.

This is an extremely simple example that doesn't _really_ need generics (it's
stringly-typed :D), but should give an idea of how functional streams can be
used.

First, let's introduce a helper function that allow us to generate a lines
stream from an `io.Reader`:

```go
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
```

> **Note:** that we're eating errors here, just another demonstration that you
> shouldn't do this in production, at least not the way it's described in this
> blog post :)
>
> The code below will also not handle any errors. Don't do this at home.

Now that we have that function, we can create our "database" instance, loop
through the input and execute commands:

```go
func ProcessCommands(input io.Reader, output io.Writer) {
	s := stream.FromReader(input)
	stream.Fold(s, NewDB(output), func(db *DB, line string) *DB {
		cmd := strings.Fields(line)
		db.Execute(cmd[0], cmd[1:])
		return db
	})
}
```

(for runnable code, checkout the [GitHub
repository](https://github.com/fsouza/blog/blob/main/content/2022-02-12-functional-streams-in-go/stream/kvdb/kvdb.go))


### Another more interesting example

A less realistic, but interesting example is taking a stream of numbers from
stdin, parse them, and sum all the prime numbers. This requires using the
helper `FromReader`, `Map` to parse the number, `Filter` to discard non-prime
numbers and `Fold` to sum the filtered values. Notice that this is stdin, so
values are getting piped through as they are read from stdin.

Here's what that "pipeline" looks like:

```go
func main() {
	stdin := stream.FromReader(os.Stdin)
	numbers := stream.Map(stdin, parseLine)
	primes := stream.Filter(numbers, isPrime)
	sum := stream.Fold(primes, 0, sum)
	fmt.Println(sum)
}
```

(again, for runnable code, checkout the GitHub repo)

## Why not methods?

**TL;DR:** Go doesn't really support it. [It may in the
future](https://github.com/golang/go/issues/43390).

One thing one may notice from the example above is that using functions makes
the code quite verbose, we have to introduce variables for intermediary streams
(or we could nest function calls).
Functional languages get away with that by having some sort of function
composition or using pipe operators / threading macros. But in more
object-oriented languages, methods are used, which is a better fit for Go. So the code below:

```go
stdin := stream.FromReader(os.Stdin)
numbers := stream.Map(stdin, parseLine)
primes := stream.Filter(numbers, isPrime)
sum := stream.Fold(primes, 0, sum)
```

Could become something like:

```go
sum := stream.FromReader(os.Stdin).Map(parseLine).Filter(isPrime).Fold(0, sum)
```

Why can't we do that in Go? The problem is that methods can't really be generic
in Go, it's not currently supported by the generics implementation, which means
the `Map` method in the pipeline above cannot be implemented. It's an issues
with how methods are used for structural subtyping & interfaces, so it may be
complicated to address or not happen at hall. See the [issue in the Go issue
tracker for more details](https://github.com/golang/go/issues/43390)!

## Feedback

Do you have any feedback? Questions? Concerns? Wanna fix a typo? Checkout the
[source for this post in
GitHub](https://github.com/fsouza/blog/blob/HEAD/content/2022-02-12-functional-streams-in-go/index.md)
(feel free to send a PR), or the [discussion in the GitHub
repo](https://github.com/fsouza/blog/discussions/12).
