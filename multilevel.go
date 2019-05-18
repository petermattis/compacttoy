package main

// The multi-level compaction strategy. A flush attempts to write data into the
// lowest non-empty level. Failing to find such a level, compactions are rooted
// at the highest level and include the subsequent levels such that the
// compaction would not cause the level being compacted into to become "too
// large". The observation behind this strategy is that if a compaction from
// Ln->Ln+1 causes Ln+1 to require a compaction, then we would have been better
// off to perform a compaction from Ln->Ln+2 (i.e. a compaction which
// encompasses Ln+1). A similar observation applies to flushes: if a flush to
// Ln triggers a compaction from Ln->Ln+1, then that compaction should be
// merged with the flush in order to reduce write amplification.
//
// Determining when a level is "too large" is done via an estimation of the
// growth in the level size due to a compaction into a level. The assumption is
// that a level will grow be the size of the input data from other levels. This
// is a pessimistic calculation that isn't accurate in the face of updates and
// deletes (data being overwritten or deleted).
//
// The inline comments below explain the math behind "too large".
func newMultiLevelStrategy(levels, target int) strategy {
	return func(s *state) {
		// When compacting level 0, allow a flush to be combined with it.
		s.levels[0]++
		for i, sum := 1, 0; i < len(s.levels); i++ {
			last := s.levels[i-1]
			sum += last

			// Write amplification is minimized when it is the same at each
			// level. The formulas below define the write amplification at each
			// level based on the size of each level.
			//
			// s(x) = size of level x
			// T(y) = sum of level sizes [1,y]
			// T(y+1) = T(y) + s(y+1)
			//
			// w(1) = s(1) / 2               ; write-amp at level 1
			// w(n) = T(n) / (T(n-1) + 1)    ; write-amp at level n

			var size int
			if i == 1 {
				// w(1) = w(2)
				//
				// s(1) / 2                       = T(2) / (T(1) + 1)
				// s(1) / 2                       = (s(1) + s(2)) / (s(1) + 1)
				// (s(1) * (s(1) + 1)) / 2        = s(1) + s(2)
				// (s(1) * (s(1) + 1)) / 2 - s(1) = s(2)
				size = (sum*(sum+1))/2 - sum
			} else {
				// w(n) = w(n+1)
				//
				// T(n) / (T(n-1) + 1)                            = T(n+1) / (T(n) + 1)
				// (T(n) * (T(n) + 1)) / (T(n-1) + 1)             = T(n+1)
				// (T(n) * (T(n) + 1)) / (T(n) - s(n) + 1)        = T(n) + s(n+1)
				// (T(n) * (T(n) + 1)) / (T(n) - s(n) + 1) - T(n) = s(n+1)
				size = (sum*(sum+1))/(sum-last+1) - sum
			}
			if size < s.levels[i] {
				s.compact(0, i-1)
				return
			}
		}
		s.compact(0, len(s.levels)-1)
	}
}
