package main

import (
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestLevenshtein(t *testing.T) {
	Terst(t)
	Is(LevenshteinDistance("a", "b"), 1)
	Is(LevenshteinDistance("a", "aa"), 1)
	Is(LevenshteinDistance("kitten", "sitting"), 3)
	Is(LevenshteinDistance("GUMBO", "GAMBOL"), 2)
}
