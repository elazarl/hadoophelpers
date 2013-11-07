package main

import (
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestLevenshtein(t *testing.T) {
	Terst(t)
	DeleteCost, ReplaceCost, AddCost = 1000, 100, 1
	Is(LevenshteinDistance("a", "b"), ReplaceCost)
	Is(LevenshteinDistance("a", "aa"), DeleteCost)
	Is(LevenshteinDistance("ab", "a"), AddCost)
	Is(LevenshteinDistance("kitten", "sitting"), 2*ReplaceCost + DeleteCost)
	Is(LevenshteinDistance("GUMBO", "GAMBOL"), ReplaceCost + DeleteCost)
}
