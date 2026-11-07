package main

import (
	"fmt"
)

func simple(x, y int) (int, int) {
	return x + y, y - x
}

func main() {
	fmt.Println(simple(1, 4))
}
