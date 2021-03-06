package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	levels = 10
	target = 16384
	update = 0.0
	unit   = 100
)

var strategies = []struct {
	name    string
	factory func(levels, target int) strategy
}{
	{"leveled", newLeveledStrategy},
	// {"tiered", newTieredStrategy},
	{"tiered+leveled", newTieredLeveledStrategy},
	{"lazy+leveled", newLazyLeveledStrategy},
	// {"flush+leveled", newFlushLeveledStrategy},
	// {"multi-level", newMultiLevelStrategy},
	{"brb-25", func(levels, target int) strategy {
		return newBRBStrategy(levels, target, 0.25)
	}},
	{"brb-33", func(levels, target int) strategy {
		return newBRBStrategy(levels, target, 0.33)
	}},
	{"brb-50", func(levels, target int) strategy {
		return newBRBStrategy(levels, target, 0.50)
	}},
	{"brb-100", func(levels, target int) strategy {
		return newBRBStrategy(levels, target, 1.0)
	}},
}

var rootCmd = &cobra.Command{
	Use:   "compacttoy [command] (flags)",
	Short: "compaction simulation toy",
	Long:  ``,
	Run:   compareStrategies,
}

// Compare the various comapction strategies. Each step of the simulation
// writes 1 unit of data, and the simulation runs for 16K steps. This is
// equivalent to writing 1 TB of data in 64 MB units.
func compareStrategies(cmd *cobra.Command, args []string) {
	target *= unit

	fmt.Printf("levels")
	for _, s := range strategies {
		fmt.Printf("%16s", s.name)
	}
	fmt.Printf("\n")
	for _, levels := range []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 50, 100} {
		fmt.Printf("%6d", levels)
		for _, s := range strategies {
			state := newState(levels)
			simulate(target, state, s.factory(levels, target))
			fmt.Printf("%16.1f", state.writeAmp())
		}
		fmt.Printf("\n")
	}
}

func main() {
	cobra.EnableCommandSorting = false

	for _, s := range strategies {
		s := s
		cmd := &cobra.Command{
			Use: s.name,
			Run: func(cmd *cobra.Command, args []string) {
				target *= unit
				state := newState(levels)
				simulate(target, state, s.factory(levels, target))
				state.dump()
			},
		}
		cmd.Flags().IntVarP(
			&levels, "levels", "l", levels, "number of levels")
		cmd.Flags().IntVarP(
			&target, "target", "t", target, "target size")
		cmd.Flags().Float64Var(
			&update, "update", update, "update fraction")
		cmd.Flags().IntVar(
			&unit, "unit", unit, "unit size")
		cmd.Flags().BoolVarP(
			&verbose, "verbose", "v", false, "verbose logging")
		rootCmd.AddCommand(cmd)
	}

	rootCmd.Flags().IntVarP(
		&target, "target", "t", target, "target size")
	rootCmd.Flags().Float64Var(
		&update, "update", update, "update fraction")
	rootCmd.Flags().IntVar(
		&unit, "unit", unit, "unit size")

	if err := rootCmd.Execute(); err != nil {
		// Cobra has already printed the error message.
		os.Exit(1)
	}
}
