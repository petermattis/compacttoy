package main

// The tiered+leveled compaction strategy. This is the strategy used by
// RocksDB. By default there are 4 tiered levels which are flushed into. When
// the tiers are all non-empty they are compacted with "level 0" and then
// compaction proceeds level-by-level from there.
//
// NB: this is not the same as the RocksDB dynamic_level_bytes configuration,
// which dynamically computes the number of levels and the max size for each
// level.
func newTieredLeveledStrategy(levels, target int) strategy {
	tiers := 4
	switch levels {
	case 2, 3, 4, 5:
		tiers = 1
	case 6, 7, 8:
		tiers = 2
	case 9, 10, 11:
		tiers = 3
	}
	maxLevelSize := makeMaxLevelSize(levels-tiers, target)

	return func(s *state) {
		// Flush a unit of data to the lowest non-empty tier.
		for i := tiers - 1; i >= 0; i-- {
			if s.levels[i] == 0 {
				s.flush(i)
				return
			}
		}

		// When all of the tiers are full, compact them to the first level.
		s.compact(0, tiers)

		// Loop over the levels, performing any compactions that are needed.
		for i := tiers; i < len(s.levels)-1; i++ {
			if s.levels[i] >= maxLevelSize[i-tiers] {
				s.compact(i, i+1)
			}
		}
	}
}
