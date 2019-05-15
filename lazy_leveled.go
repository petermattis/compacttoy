package main

// The lazy-leveled compaction strategy as described in:
//  https://stratos.seas.harvard.edu/files/stratos/files/dostoevskykv.pdf
//
// The upper-levels have multiple tiers, while the bottom level has a single
// tier. When all of the tiers in a multi-tier level become full, they are
// compacted into the lowest empty tier in the next level.
func newLazyLeveledStrategy(levels, target int) strategy {
	tiers := 4
	switch levels {
	case 2:
		tiers = 1
	case 3, 4, 5:
		tiers = 2
	case 6, 7, 8:
		tiers = 3
	case 9, 10, 11:
		tiers = 3
	}

	tiersPerLevel := make([]int, (levels-1+tiers-1)/tiers)
	tiersPerLevel[0] = levels - 1
	for i := 1; i < len(tiersPerLevel); i++ {
		tiersPerLevel[i] = tiers
		tiersPerLevel[0] -= tiers
	}

	return func(s *state) {
		// Flush a unit of data to the lowest non-empty tier.
		for i := tiersPerLevel[0] - 1; i >= 0; i-- {
			if s.levels[i] == 0 {
				s.flush(i)
				return
			}
		}

		// When all of the tiers for a level are full, compact them to the first
		// non-empty tier in the next level.
		start := 0
		for j := 1; j < len(tiersPerLevel); j++ {
			base := start + tiersPerLevel[j-1]
			for i := tiersPerLevel[j] - 1; i >= 0; i-- {
				if s.levels[base+i] == 0 {
					s.compact(start, base+i)
					if i != 0 {
						return
					}
					break
				}
			}
			start = base
		}

		// All of the tiers in the last level are full. Compact them in order to
		// make room for future compactions.
		s.compact(levels-tiersPerLevel[len(tiersPerLevel)-1]-1, levels-1)
	}
}
