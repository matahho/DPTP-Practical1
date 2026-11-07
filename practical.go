// Filename: practical.go
// Usage: $ go run proc.go practical.go
// Encodings of polyadic synchronous pi into monadic asynchronous pi
// Syntax of processes do not contain base values: only channels are passed around
// GPLv3 2025, MobilityReadingGroup, University of Oxford

package main

import (
	"fmt"
	"strconv"
)

var free_name_counter = 0
var free_var_counter = 0

func fresh_name() string {
	new_name := "_a" + strconv.Itoa(free_name_counter)
	free_name_counter++
	return new_name
}

func fresh_var() string {
	new_var := "_x" + strconv.Itoa(free_var_counter)
	free_var_counter++
	return new_var
}

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

	case parallel_proc:
		return parallel_proc{
			left:  to_monadic(p.left),
			right: to_monadic(p.right),
		}
	case repl_proc:
		return repl_proc{
			to_monadic(p.cont),
		}

	case input_proc:
		stablished_channel := fresh_var()
		var constructed_cont proc
		constructed_cont = to_monadic(p.cont)

		for i := len(p.vars) - 1; i >= 0; i-- {
			generated := input_proc{
				channel: stablished_channel,
				vars:    []string{p.vars[i]},
				cont:    constructed_cont,
			}
			constructed_cont = generated
		}

		return input_proc{
			channel: p.channel,
			vars:    []string{stablished_channel},
			cont:    constructed_cont,
		}

	case output_proc:
		stablished_channel := fresh_name()
		var constructed_cont proc
		constructed_cont = to_monadic(p.cont)
		for i := len(p.values) - 1; i >= 0; i-- {
			generated := output_proc{
				channel: stablished_channel,
				values:  []string{p.values[i]},
				cont:    constructed_cont,
			}
			constructed_cont = generated
		}

		return res_proc{
			channel: stablished_channel,
			cont: output_proc{
				channel: p.channel,
				values:  []string{stablished_channel},
				cont:    constructed_cont,
			},
		}

	}
	return nil_proc{}
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
	p := example3()
	fmt.Printf("Example: %s\n\n", p)
	fmt.Println(to_monadic(p))
	fmt.Println(to_monadic(p).is_monadic())

	//
	//pi := indirect_encoding(p)
	//fmt.Printf("Indirect Encoding: %s\n\n", pi)
	//pd := direct_encoding(p)
	//fmt.Printf("Direct Encoding: %s\n\n", pd)
	//run(pi, make(map[string]chan any))
	//run(pd, make(map[string]chan any))
}
