// Filename: practical.go
// Usage: $ go run proc.go practical.go
// Encodings of polyadic synchronous pi into monadic asynchronous pi
// Syntax of processes do not contain base values: only channels are passed around
// GPLv3 2025, MobilityReadingGroup, University of Oxford

// The reason that it need less messages/channels is that because it needs less new channel establishment.
// in the indirect encoding for each variable/name we need to establish a new channel for sending each pair
// but in the direct encoding we only establish a channel once then we can use it for all of our data as well.
// Also in the indirect encoding we need some new channels/messages for both monadic transition and async transition seperately

package main

import (
	"fmt"
	"maps"
	"strconv"
	"time"
)

type CommunicationEfficiency struct {
	numberOfChannels int
	numberOfMessages int
}

var directEncodingEfficiency CommunicationEfficiency
var indirectEncodingEfficiency CommunicationEfficiency

var freeNameCounter = 0
var freeVarCounter = 0

func fresh_name() string {
	newName := "_a" + strconv.Itoa(freeNameCounter)
	freeNameCounter++
	return newName
}

func fresh_var() string {
	newVar := "_x" + strconv.Itoa(freeVarCounter)
	freeVarCounter++
	return newVar
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
		if (len(p.vars)) == 1 {
			return p
		}

		established_channel := fresh_var()
		var constructed_cont proc
		constructed_cont = to_monadic(p.cont)

		for i := len(p.vars) - 1; i >= 0; i-- {
			generated := input_proc{
				channel: established_channel,
				vars:    []string{p.vars[i]},
				cont:    constructed_cont,
			}
			constructed_cont = generated
		}

		return input_proc{
			channel: p.channel,
			vars:    []string{established_channel},
			cont:    constructed_cont,
		}

	case output_proc:
		if len(p.values) == 1 {
			return p
		}

		established_channel := fresh_name()
		var constructed_cont proc
		constructed_cont = to_monadic(p.cont)
		for i := len(p.values) - 1; i >= 0; i-- {
			generated := output_proc{
				channel: established_channel,
				values:  []string{p.values[i]},
				cont:    constructed_cont,
			}
			constructed_cont = generated
		}

		return res_proc{
			channel: established_channel,
			cont: output_proc{
				channel: p.channel,
				values:  []string{established_channel},
				cont:    constructed_cont,
			},
		}
	case nil:
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

	case parallel_proc:
		return parallel_proc{
			left:  to_async(p.left),
			right: to_async(p.right),
		}

	case repl_proc:
		return repl_proc{
			cont: to_async(p.cont),
		}

	case input_proc:

		established_channel_on_receiver_side := fresh_name() // which is "d" in the doc
		established_channel_on_sender_side := fresh_var()    // which is "y" in the doc

		return input_proc{
			channel: p.channel,
			vars:    []string{established_channel_on_sender_side},
			cont: res_proc{
				channel: established_channel_on_receiver_side,
				cont: parallel_proc{
					output_proc{
						established_channel_on_sender_side,
						[]string{established_channel_on_receiver_side},
						nil,
					},
					input_proc{
						established_channel_on_receiver_side,
						[]string{p.vars[0]}, // we are sure that it is monadic (because of the panic in the beginning of this func)
						to_async(p.cont),
					},
				},
			},
		}

	case output_proc:
		established_channel_on_receiver_side := fresh_var() // which is "y" in the doc
		established_channel_on_sender_side := fresh_name()  // which is "c" in the doc
		return res_proc{
			channel: established_channel_on_sender_side,
			cont: parallel_proc{
				output_proc{
					channel: p.channel,
					values:  []string{established_channel_on_sender_side},
					cont:    nil,
				},
				input_proc{
					channel: established_channel_on_sender_side,
					vars:    []string{established_channel_on_receiver_side},
					cont: parallel_proc{
						left: output_proc{
							channel: established_channel_on_receiver_side,
							values:  []string{p.values[0]}, // we are sure that it is monadic (because of the panic in the beginning of this func)
							cont:    nil,
						},
						right: to_async(p.cont),
					},
				},
			},
		}
	case nil:
		return nil_proc{}
	default:
		panic("Not implemented")

	}
}

func indirect_encoding(p proc) proc {
	var monadic_encoding = to_monadic(p)
	var monadic_async_encoding = to_async(monadic_encoding)
	return monadic_async_encoding
}

func output_direct_encoding(p output_proc) proc {
	var fixed_established_channel = fresh_name()
	var continuation = direct_encoding(p.cont)

	for i := len(p.values) - 1; i >= 0; i-- {
		var temp_channel = fresh_var()

		var sending_data = output_proc{
			temp_channel,
			[]string{p.values[i]},
			nil,
		}

		continuation = parallel_proc{
			right: continuation,
			left:  sending_data,
		}

		continuation = input_proc{
			channel: fixed_established_channel,
			vars:    []string{temp_channel},
			cont:    continuation,
		}

	}
	return res_proc{
		channel: fixed_established_channel,
		cont: parallel_proc{
			output_proc{
				p.channel,
				[]string{fixed_established_channel},
				nil,
			},
			continuation,
		},
	}
}

func input_direct_encoding(p input_proc) proc {
	var received_channel = fresh_var() // in here it is "y"
	var main_channel = fresh_name()    // in here it is "d"
	var continuation = direct_encoding(p.cont)

	for i := len(p.vars) - 1; i >= 0; i-- {

		sending_channel_to_receive_data := output_proc{
			channel: received_channel,
			values:  []string{main_channel},
			cont:    nil,
		}
		receiving_message := input_proc{
			main_channel,
			[]string{p.vars[i]},
			continuation,
		}

		continuation = parallel_proc{
			left:  sending_channel_to_receive_data,
			right: receiving_message,
		}
	}

	return input_proc{
		channel: p.channel,
		vars:    []string{received_channel},
		cont: res_proc{
			main_channel,
			continuation,
		},
	}
}

func direct_encoding(p proc) proc {
	switch p := p.(type) {
	case nil_proc:
		return nil_proc{}

	case parallel_proc:
		return parallel_proc{
			left:  direct_encoding(p.left),
			right: direct_encoding(p.right),
		}

	case input_proc:
		return input_direct_encoding(p)
	case output_proc:
		return output_direct_encoding(p)

	case res_proc:
		return res_proc{
			channel: p.channel,
			cont:    direct_encoding(p.cont),
		}
	case repl_proc:
		return repl_proc{
			direct_encoding(p.cont),
		}

	case nil:
		return nil_proc{}
	default:
		panic("Not implemented")
	}

}

const chan_buffer = 8

func lookup(stack map[string]chan any, name string) chan any {
	ch, ok := stack[name]
	if !ok {
		panic("unbound name or variable: " + name)
	}
	return ch
}

// Everything travelling over a channel is itself a channel, so every received value must be a chan any.
func as_channel(val any, where string) chan any {
	ch, ok := val.(chan any)
	if !ok {
		panic("received a non-channel value on " + where)
	}
	return ch
}

// runner for monadic asynchronous pi-calculus (might not terminate)
func run(p proc, stack map[string]chan any) {
	if !p.is_monadic_async() {
		panic("Input process must be monadic and asynchronous")
	}

	switch p := p.(type) {

	case nil_proc:
		return

	case res_proc:
		stack[p.channel] = make(chan any, chan_buffer)
		fmt.Printf("[nu]   %s\n", p.channel)
		run(p.cont, stack)

	case output_proc:
		// u < a > , since it is async no more cont is needed
		target := lookup(stack, p.channel)
		payload := lookup(stack, p.values[0])
		fmt.Printf("[send] %s<%s>\n", p.channel, p.values[0])
		target <- payload
		return

	case input_proc:
		// u(x).P : block until a channel arrives on u
		source := lookup(stack, p.channel) // u
		val := <-source                    // x
		stack[p.vars[0]] = as_channel(val, p.channel)
		fmt.Printf("[recv] %s(%s)\n", p.channel, p.vars[0])
		run(p.cont, stack)

	case parallel_proc:
		go run(p.left, maps.Clone(stack))
		go run(p.right, maps.Clone(stack))
		return

	case repl_proc:
		guard, ok := p.cont.(input_proc)
		if !ok {
			panic("replication must be input-guarded")
		}
		source := lookup(stack, guard.channel)
		base := maps.Clone(stack)
		go func() {
			for {
				val := <-source
				s := maps.Clone(base)
				s[guard.vars[0]] = as_channel(val, guard.channel)
				fmt.Printf("[recv] %s(%s) (replicated)\n", guard.channel, guard.vars[0])
				go run(guard.cont, s)
			}
		}()
		return

	default:
		panic("Not implemented")
	}
}

func main() {
	p := example3()
	fmt.Printf("Example: %s\n\n", p)

	pi := indirect_encoding(p)
	fmt.Printf("Indirect Encoding: %s\n\n", pi)

	pd := direct_encoding(p)
	fmt.Printf("Direct Encoding: %s\n\n", pd)

	fmt.Println("--- running indirect encoding ---")
	run(pi, make(map[string]chan any))
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n--- running direct encoding ---")
	run(pd, make(map[string]chan any))
	time.Sleep(500 * time.Millisecond)
}
