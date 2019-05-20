package main

import "fmt"

var (
	verbose = false
)

type state struct {
	levels  []int
	written []int
	flushed int
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
	var maxSize int
	for i := range s.levels {
		if maxSize < s.levels[i] {
			maxSize = s.levels[i]
		}
	}
	return 1.0 - float64(maxSize)/float64(s.totalSize())
}

func (s *state) flush(level int) {
	s.flushed += unit
	s.levels[level] += unit
	s.compact(level, level)
}

func (s *state) compact(start, output int) int {
	var in int
	var nonEmptyLevels int
	for i := start; i < output; i++ {
		if s.levels[i] != 0 {
			nonEmptyLevels++
		}
		in += s.levels[i]
		s.levels[i] = 0
	}

	// Model move compactions: if there is only one input level and the output
	// level is empty, we can move the inputs directly to the outputs without
	// rewriting. Because there is no rewriting, the size of the "output" is
	// exactly the size of the inputs.
	if start != output && nonEmptyLevels == 1 && s.levels[output] == 0 {
		s.levels[output] = in
		return in
	}

	sum := s.levels[output] + int(float64(in)*(1-update))
	s.levels[output] = sum
	s.written[output] += sum
	return sum
}

func (s *state) dump() {
	fmt.Printf("level      size     write     space\n")
	total := s.totalSize()
	for i := range s.levels {
		fmt.Printf("%5d %9d %9d %8.1f%%\n", i, s.levels[i], s.written[i],
			100.0*float64(s.levels[i])/float64(total))
	}
	fmt.Printf("total %9d %9d\n", total, s.totalWritten())
	fmt.Printf("w-amp %9.1f\n", s.writeAmp())
	fmt.Printf("s-amp %8.1f%%\n", 100.0*s.spaceAmp())
	fmt.Printf("\n")
}

type strategy func(s *state)

func simulate(target int, state *state, strategy strategy) {
	for {
		if state.flushed >= target {
			break
		}
		strategy(state)
		if verbose {
			state.dump()
		}
	}
}
