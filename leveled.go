package main

import (
	"math"
)

func makeMaxLevelSize(levels, target int) []int {
	// NB: poor-man's iteration to find the per-level multiplier. This could be
	// more sophisticated (i.e. newton-raphson iteration), but this suffices
	// for now.
	m := 1.1
	for {
		// 1 + x + x^2 + x^3 + ... + x^n = (1 - x^(n+1)) / (1-x)
		size := int((1 - math.Pow(m, float64(levels+1))) / (1 - m))
		if size >= target {
			break
		}
		m += 0.1
	}

	max := make([]int, levels)
	for i, base := 0, float64(1); i < len(max); i++ {
		max[i] = int(base)
		base *= m
	}
	return max
}

// The classic leveled compaction strategy. Flushing always goes to level 0 and
// that level is required to be empty to flush. Compaction proceeds
// level-by-level after that, whenever a level is larger than its max size.
func newLeveledStrategy(levels, target int) strategy {
	maxLevelSize := makeMaxLevelSize(levels, target)

	return func(s *state) {
		if s.levels[0] == 0 {
			// L0 is empty. Flush to it.
			s.flush(0)
			return
		}

		// Loop over the levels, performing any compactions that are needed.
		for i := 0; i < len(s.levels)-1; i++ {
			if s.levels[i] >= maxLevelSize[i] {
				end := i + 1
				for ; end < len(s.levels)-1; end++ {
					if s.levels[end+1] != 0 {
						break
					}
				}
				s.compact(i, end)
			}
		}
	}
}

// A variant of the classic leveled compaction strategy. Flushing is allowed to
// merge with the first level, and the first level is allowed to be
// significantly larger in size.
func newFlushLeveledStrategy(levels, target int) strategy {
	maxLevelSize := makeMaxLevelSize(levels, target)
	// Because we're willing to flush and compact at the same time, the target
	// size of the first level can be larger.
	copy(maxLevelSize, maxLevelSize[1:])

	return func(s *state) {
		if s.levels[0] == 0 {
			// L0 is empty. Flush to it.
			s.flush(0)
			return
		}

		// Loop over the levels, performing any compactions that are
		// needed. Similar to the leveled strategy, except that we're willing to
		// perform a flush+compact operation to the first level.
		flushed := false
		for i := 0; i < len(s.levels)-1; i++ {
			if s.levels[i] >= maxLevelSize[i] {
				if i == 0 {
					flushed = true
					s.levels[i] += unit
				}
				end := i + 1
				for ; end < len(s.levels)-1; end++ {
					if s.levels[end+1] != 0 {
						break
					}
				}
				s.compact(i, end)
			}
		}
		if !flushed {
			s.flush(0)
		}
	}
}
