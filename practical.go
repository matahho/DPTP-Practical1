// Filename: practical.go
// Usage: $ go run proc.go practical.go
// Encodings of polyadic synchronous pi into monadic asynchronous pi
// Syntax of processes do not contain base values: only channels are passed around
// GPLv3 2025, MobilityReadingGroup, University of Oxford

package main

import "fmt"

// polyadic synchronous to monadic synchronous
func to_monadic(p proc) proc {
	switch p := p.(type) {

	case res_proc:
		return res_proc{
			channel: p.channel,
			cont:    to_monadic(p.cont),
		}

	case nil_proc:
		return nil_proc{}

	default:
		panic("Not implemented")
	}
}

// monadic synchronous to monadic asynchronous
func to_async(p proc) proc {
	if !p.is_monadic() {
		panic("Input process must be monadic")
	}

	switch p := p.(type) {

	case res_proc:
		return res_proc{
			channel: p.channel,
			cont:    to_async(p.cont),
		}

	case nil_proc:
		return nil_proc{}

	default:
		panic("Not implemented")
	}
}

func indirect_encoding(p proc) proc {
	panic("Not implemented")
}

func direct_encoding(p proc) proc {
	panic("Not implemented")
}

// runner for monadic asynchronous pi-calculus (might not terminate)
func run(p proc, stack map[string]chan any) {
	if !p.is_monadic_async() {
		panic("Input process must be monadic and asynchronous")
	}
	panic("Not implemented")
}

func main() {
	p := example1()
	fmt.Printf("Example: %s\n\n", p)
	pi := indirect_encoding(p)
	fmt.Printf("Indirect Encoding: %s\n\n", pi)
	pd := direct_encoding(p)
	fmt.Printf("Direct Encoding: %s\n\n", pd)
	run(pi, make(map[string]chan any))
	run(pd, make(map[string]chan any))
}
