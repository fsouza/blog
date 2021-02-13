+++
title = "Go solution for the Dining philosophers problem"
aliases = ["/2011/10/go-solution-for-dining-philosophers.html"]

[taxonomies]
tags = ["concurrency", "go", "golang"]
+++

I spent part of the Sunday solving the [Dining
Philosophers](https://en.wikipedia.org/wiki/Dining_philosophers_problem) using
[Go](https://golang.org/). The given solution is based in the description for
the problem present in [The Little Book of
Semaphores](http://greenteapress.com/semaphores/):

> The Dining Philosophers Problem was proposed by Dijkstra in 1965, when
> dinosaurs ruled the earth. It appears in a number of variations, but the
> standard features are a table with ﬁve plates, ﬁve forks (or chopsticks) and a
> big bowl of spaghetti.

There are some constraints:

- Only one philosopher can hold a fork at a time
- It must be impossible for a deadlock to occur
- It must be impossible for a philosopher to starve waiting for a fork
- It must be possible for more than one philosopher to eat at the same time

No more talk, here is my solution for the problem:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Fork struct {
	sync.Mutex
}

type Table struct {
	philosophers chan Philosopher
	forks []*Fork
}

func NewTable(forks int) *Table {
	t := new(Table)
	t.philosophers = make(chan Philosopher, forks - 1)
	t.forks = make([]*Fork, forks)
	for i := 0; i < forks; i++ {
		t.forks[i] = new(Fork)
	}
	return t
}

func (t *Table) PushPhilosopher(p Philosopher) {
	p.table = t
	t.philosophers <- p
}

func (t *Table) PopPhilosopher() Philosopher {
	p := <-t.philosophers
	p.table = nil
	return p
}

func (t *Table) RightFork(philosopherIndex int) *Fork {
	f := t.forks[philosopherIndex]
	return f
}

func (t *Table) LeftFork(philosopherIndex int) *Fork {
	f := t.forks[(philosopherIndex + 1) % len(t.forks)]
	return f
}

type Philosopher struct {
	name string
	index int
	table *Table
	fed chan int
}

func (p Philosopher) Think() {
	fmt.Printf("%s is thinking...\n", p.name)
	time.Sleep(3e9)
	p.table.PushPhilosopher(p)
}

func (p Philosopher) Eat() {
	p.GetForks()
	fmt.Printf("%s is eating...\n", p.name)
	time.Sleep(3e9)
	p.PutForks()
	p.table.PopPhilosopher()
	p.fed <- 1
}

func (p Philosopher) GetForks() {
	rightFork := p.table.RightFork(p.index)
	rightFork.Lock()

	leftFork := p.table.LeftFork(p.index)
	leftFork.Lock()
}

func (p Philosopher) PutForks() {
	rightFork := p.table.RightFork(p.index)
	rightFork.Unlock()

	leftFork := p.table.LeftFork(p.index)
	leftFork.Unlock()
}

func main() {
	table := NewTable(5)
	philosophers := []Philosopher{
		Philosopher{"Thomas Nagel", 0, table, make(chan int)},
		Philosopher{"Elizabeth Anscombe", 1, table, make(chan int)},
		Philosopher{"Martin Heidegger", 2, table, make(chan int)},
		Philosopher{"Peter Lombard", 3, table, make(chan int)},
		Philosopher{"Gottfried Leibniz", 4, table, make(chan int)},
	}

	for {
		for _, p := range philosophers {
			go func(p Philosopher){
				p.Think()
				p.Eat()
			}(p)
		}

		for _, p := range philosophers {
			<-p.fed
			fmt.Printf("%s was fed.\n", p.name)
		}
	}

}
```
