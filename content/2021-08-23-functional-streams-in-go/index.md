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
