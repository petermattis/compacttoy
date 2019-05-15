package main

// Tiered compaction strategy. There are ceil(levels/tiers) tier-levels. All
// but the first tier-level has 4 tiers. The first tier-level has between [1,4]
// tiers. Tiers are populated "all to all": when all of the tiers in a
// tier-level are non-empty, they are compacted to the lowest non-empty tier in
// the next tier-level. If all of the tiers in the bottom tier-level are full,
// they are compacted to the lowest tier in the bottom tier-level.
func newTieredStrategy(levels, target int) strategy {
	tiers := 4
	if tiers >= levels {
		tiers = levels - 1
	}

	tiersPerLevel := make([]int, (levels+tiers-1)/tiers)
	tiersPerLevel[0] = levels
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
		s.compact(levels-tiers-1, levels-1)
	}
}
