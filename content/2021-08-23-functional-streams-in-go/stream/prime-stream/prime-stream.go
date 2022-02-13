package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fsouza/stream"
)

func main() {
	stdin := stream.FromReader(os.Stdin)
	numbers := stream.Map(stdin, parseLine)
	primes := stream.Filter(numbers, isPrime)
	sum := stream.Fold(primes, 0, sum)
	fmt.Println(sum)
}

func sum(x, y int) int {
	return x + y
}

func parseLine(line string) int {
	v, err := strconv.Atoi(line)
	if err != nil {
		return 0
	}
	return v
}

func isPrime(v int) bool {
	if v < 2 {
		return false
	}
	for i := 2; i < v; i++ {
		if v%i == 0 {
			return false
		}
	}
	return true
}
