// Filename: proc.go
// Usage: $ go run proc.go practical.go
// GPLv3 2025, MobilityReadingGroup, University of Oxford

package main

import (
	"fmt"
	"strings"
)

// Algebraic types representing polyadic or monadic, synchronous or asynchronous pi calculus processes.
//
// Monadic: the number of vars in input_proc and the number of vars in output_proc is exactly 1
//
// Async: output_proc has no continuation (cont = nil)
type proc interface {
	is_monadic() bool
	is_monadic_async() bool
}

type input_proc struct {
	channel string
	vars    []string
	cont    proc
}

func (p input_proc) is_monadic() bool {
	return (len(p.vars) == 1) && p.cont.is_monadic()
}
func (p input_proc) is_monadic_async() bool {
	return (len(p.vars) == 1) && p.cont.is_monadic_async()
}
func (p input_proc) String() string {
	return fmt.Sprintf("%s(%s).%s", p.channel, strings.Join(p.vars, ","), p.cont)
}

type output_proc struct {
	channel string
	values  []string
	cont    proc
}

func (p output_proc) is_monadic() bool {
	return (len(p.values) == 1) && p.cont.is_monadic()
}
func (p output_proc) is_monadic_async() bool {
	return (len(p.values) == 1) && p.cont == nil // async pi-calculus does not have continuation
} // Ask this if nil == nil_proc or not?
func (p output_proc) String() string {
	if p.cont == nil {
		return fmt.Sprintf("-%s<%s>", p.channel, strings.Join(p.values, ","))
	}
	return fmt.Sprintf("-%s<%s>.%s", p.channel, strings.Join(p.values, ","), p.cont)
}

type parallel_proc struct {
	left, right proc
}

func (p parallel_proc) is_monadic() bool {
	return p.left.is_monadic() && p.right.is_monadic()
}
func (p parallel_proc) is_monadic_async() bool {
	return p.left.is_monadic_async() && p.right.is_monadic_async()
}
func (p parallel_proc) String() string {
	return fmt.Sprintf("(%s | %s)", p.left, p.right)
}

type repl_proc struct {
	cont proc
}

func (p repl_proc) is_monadic() bool {
	return p.cont.is_monadic()
}
func (p repl_proc) is_monadic_async() bool {
	return p.cont.is_monadic_async()
}
func (p repl_proc) String() string {
	return fmt.Sprintf("!%s", p.cont)
}

type res_proc struct {
	channel string
	cont    proc
}

func (p res_proc) is_monadic() bool {
	return p.cont.is_monadic()
}
func (p res_proc) is_monadic_async() bool {
	return p.cont.is_monadic_async()
}
func (p res_proc) String() string {
	return fmt.Sprintf("(nu %s)%s", p.channel, p.cont)
}

type nil_proc struct{}

func (_ nil_proc) is_monadic() bool {
	return true
}
func (_ nil_proc) is_monadic_async() bool {
	return true
}
func (p nil_proc) String() string {
	return "0"
}

// Examples of polyadic synchronous processes

// (nu a b1 b2 b3)(a<b1,b2,b3>.b2(y) | a(x1,x2,x3).(ve)x2<e>)
func example1() proc {
	// a<b1,b2,b3>.b2(y)
	p1 := output_proc{
		channel: "a",
		values:  []string{"b1", "b2", "b3"},
		cont: input_proc{
			channel: "b2",
			vars:    []string{"y"},
			cont:    nil_proc{},
		},
	}

	// a(x1,x2,x3).(ve)x2<e>
	p2 := input_proc{
		channel: "a",
		vars:    []string{"x1", "x2", "x3"},
		cont: res_proc{
			channel: "e",
			cont: output_proc{
				channel: "x2",
				values:  []string{"e"},
				cont:    nil_proc{},
			},
		},
	}

	return res_proc{channel: "a", cont: res_proc{channel: "b1", cont: res_proc{channel: "b2", cont: res_proc{channel: "b3",
		cont: parallel_proc{
			left:  p1,
			right: p2,
		}}}}}
}

// (nu a b1 b2 b3)(a<b1,b2,b3>.b2(y) | a(x1,x2,x3).b2<x3>)
func example2() proc {
	// a<b1,b2,b3>.b2(y)
	p1 := output_proc{
		channel: "a",
		values:  []string{"b1", "b2", "b3"},
		cont: input_proc{
			channel: "b2",
			vars:    []string{"y"},
			cont:    nil_proc{},
		},
	}

	// a(x1,x2,x3).b2<x3>
	p2 := input_proc{
		channel: "a",
		vars:    []string{"x1", "x2", "x3"},
		cont: output_proc{
			channel: "b2",
			values:  []string{"x3"},
			cont:    nil_proc{},
		},
	}

	return res_proc{channel: "a", cont: res_proc{channel: "b1", cont: res_proc{channel: "b2", cont: res_proc{channel: "b3",
		cont: parallel_proc{
			left:  p1,
			right: p2,
		}}}}}
}

// (nu a b c d)(!a(x1,x2,x3).x1<x3>.(ve)x2<e> | !b(x).c<x> | !c(x).0 | a<b,c,d> | a<b,c,d> | a<c,b,d> | a<c,b,d>)
func example3() proc {
	// !a(x1,x2,x3).x1<x3>.(ve)x2<e>
	p1 := repl_proc{
		cont: input_proc{
			channel: "a",
			vars:    []string{"x1", "x2", "x3"},
			cont: output_proc{
				channel: "x1",
				values:  []string{"x3"},
				cont: res_proc{
					channel: "e",
					cont: output_proc{
						channel: "x2",
						values:  []string{"e"},
						cont:    nil_proc{},
					},
				},
			},
		},
	}

	// !b(x).c<x>
	p2 := repl_proc{
		cont: input_proc{
			channel: "b",
			vars:    []string{"x"},
			cont: output_proc{
				channel: "c",
				values:  []string{"x"},
				cont:    nil_proc{},
			},
		},
	}

	// !c(x).0
	p3 := repl_proc{
		cont: input_proc{
			channel: "c",
			vars:    []string{"x"},
			cont:    nil_proc{},
		},
	}

	// a<b,c,d>
	p4 := output_proc{
		channel: "a",
		values:  []string{"b", "c", "d"},
		cont:    nil_proc{},
	}

	// a<c,b,d>
	p5 := output_proc{
		channel: "a",
		values:  []string{"c", "b", "d"},
		cont:    nil_proc{},
	}

	return res_proc{channel: "a", cont: res_proc{channel: "b", cont: res_proc{channel: "c", cont: res_proc{channel: "d",
		cont: parallel_proc{
			left: p1,
			right: parallel_proc{
				left: p2,
				right: parallel_proc{
					left: p3,
					right: parallel_proc{
						left: p4,
						right: parallel_proc{
							left: p4,
							right: parallel_proc{
								left:  p5,
								right: p5,
							}}}}}}}}}}
}

func my_example1() proc {
	// a(x1,x2,x3)
	p1 := input_proc{
		channel: "a",
		vars:    []string{"x1", "x2", "x3"},
		cont:    nil_proc{},
	}
	return p1
}

func my_example2() proc {
	// a(x1,x2,x3).b(y1,y2,y3).0
	p1 := input_proc{
		channel: "a",
		vars:    []string{"x1", "x2", "x3"},
		cont: input_proc{
			channel: "b",
			vars:    []string{"y1", "y2", "y3"},
			cont:    nil_proc{},
		},
	}
	return p1
}

func my_example3() proc {
	// a<b1>
	p1 := output_proc{
		channel: "a",
		values:  []string{"b1"},
		cont:    nil_proc{},
	}
	return p1
}

func output_polyadic_sync_example() proc {
	return output_proc{
		channel: "u",
		values:  []string{"v1", "v2", "v3"},
		cont:    nil,
	}
}

func input_polyadic_sync_example() proc {
	return input_proc{
		channel: "u",
		vars:    []string{"x1", "x2", "x3"},
		cont:    nil,
	}
}
