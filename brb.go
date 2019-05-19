package main

// The balanced rent-or-buy compaction strategy as detailed in:
//
//   https://arxiv.org/pdf/1407.3008.pdf
//
// See also this presentation:
//
//   http://www.cs.ucr.edu/~neal/Slides/bigtable_merge_compaction.pdf
//
// Note that the paper and presentation only give very rough guidance for this
// compaction strategy. The implementation below might be incorrect.
func newBRBStrategy(levels, target int) strategy {
	type levelInfo struct {
		// The cost (size) of the last compaction to this level.
		lastCost int
		// Snapshot of the "written" state after the last compaction. Used to
		// compute the cost of the compactions for higher levels since the last
		// compaction.
		lastWritten []int
	}

	info := make([]levelInfo, levels)
	for i := range info {
		info[i].lastWritten = make([]int, levels)
	}

	inCost := func(s *state, level int) int {
		var c int
		for i := 0; i < level; i++ {
			c += s.written[i] - info[level].lastWritten[i]
		}
		return c
	}

	return func(s *state) {
		// Always flush to the highest level. This doesn't actually account for the
		// write to disk. The loop below will include the flush data with any
		// necessary compaction.
		s.levels[0] += unit
		s.flushed += unit
		for i := 1; i <= len(s.levels); i++ {
			// Include level i in the compaction when the cost of compactions at
			// higher levels (inCost) >= cost of previous compaction to level times
			// the level number (i).
			if i == len(s.levels) || inCost(s, i) < i*info[i].lastCost {
				size := s.compact(0, i-1)
				info[i-1].lastCost = size
				copy(info[i-1].lastWritten, s.written)
				return
			}
		}
	}
}
