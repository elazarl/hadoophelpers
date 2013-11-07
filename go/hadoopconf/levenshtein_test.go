package main

import (
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestLevenshtein(t *testing.T) {
	Terst(t)
	DeleteCost, ReplaceCost, AddCost = 1000, 100, 1
	// note, har -> ha[doop.tmp.dir]r
	Is(LevenshteinDistance("hadoop.tmp.dir", "har"), AddCost*11)
	Is(LevenshteinDistance("fs.har.impl.disable.cache", "har"), AddCost*(3+19))
	Is(LevenshteinDistance("a", "b"), ReplaceCost)
	Is(LevenshteinDistance("a", "aa"), DeleteCost)
	Is(LevenshteinDistance("ab", "a"), AddCost)
	Is(LevenshteinDistance("kitten", "sitting"), 2*ReplaceCost + DeleteCost)
	Is(LevenshteinDistance("GUMBO", "GAMBOL"), ReplaceCost + DeleteCost)
}
