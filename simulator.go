package main

import "fmt"

var (
	verbose = false
)

type state struct {
	levels  []int
	written []int
}

func newState(n int) *state {
	return &state{
		levels:  make([]int, n),
		written: make([]int, n),
	}
}

func (s *state) totalSize() int {
	var v int
	for i := range s.levels {
		v += s.levels[i]
	}
	return v
}

func (s *state) totalWritten() int {
	var v int
	for i := range s.written {
		v += s.written[i]
	}
	return v
}

func (s *state) writeAmp() float64 {
	return float64(s.totalWritten()) / float64(s.totalSize())
}

func (s *state) spaceAmp() float64 {
	return 1.0 - float64(s.levels[len(s.levels)-1])/float64(s.totalSize())
}

func (s *state) flush(level int) {
	s.levels[level] += unit
	s.compact(level, level)
}

func (s *state) compact(start, output int) int {
	const updateFraction = 0.0

	var in int
	for i := start; i < output; i++ {
		in += s.levels[i]
		s.levels[i] = 0
	}
	sum := s.levels[output] + int(float64(in)*(1-updateFraction))
	s.levels[output] = sum
	s.written[output] += sum
	return sum
}

func (s *state) dump() {
	fmt.Printf("level    size   write   space\n")
	total := s.totalSize()
	for i := range s.levels {
		fmt.Printf("%5d %7d %7d %6.1f%%\n", i, s.levels[i], s.written[i],
			100.0*float64(s.levels[i])/float64(total))
	}
	fmt.Printf("total %7d %7d\n", total, s.totalWritten())
	fmt.Printf("w-amp %7.1f\n", s.writeAmp())
	fmt.Printf("\n")
}

type strategy func(s *state)

func simulate(target int, state *state, strategy strategy) {
	for {
		if state.totalSize() >= target {
			break
		}
		strategy(state)
		if verbose {
			state.dump()
		}
	}
}
