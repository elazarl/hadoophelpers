package main

import (
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
)

type weightedStrings struct {
	strings []string
	weights []int
}

func (w *weightedStrings) Swap(i, j int) {
	w.strings[i], w.strings[j] = w.strings[j], w.strings[i]
	w.weights[i], w.weights[j] = w.weights[j], w.weights[i]
}

func (w *weightedStrings) Less(i, j int) bool {
	return w.weights[i] < w.weights[j]
}

func (w *weightedStrings) Len() int {
	return len(w.strings)
}

func (w *weightedStrings) Append(s string, weight int) {
	w.strings = append(w.strings, s)
	w.weights = append(w.weights, weight)
}

func fuzzyScore(candidate, idol string) int {
	d := LevenshteinDistance(candidate, idol)
	if strings.HasPrefix(idol, candidate) {
		d -= 1000 // 1,000 is the infininty of the levenstein distance
	}
	return d
}

func fuzzyFind(partial string, options []string) []string {
	ws := weightedStrings{}
	for _, option := range options {
		ws.Append(option, fuzzyScore(partial, option))
	}
	sort.Sort(&ws)
	return ws.strings
}

func Complete(parser *flags.Parser, args []string) []string {
	parser.Options |= flags.IgnoreUnknown
	parser.ParseArgs(args)
	options := []string{}
	for _, group := range parser.Groups {
		for name := range group.LongNames {
			options = append(options, "--" + name)
		}
		for name := range group.ShortNames {
			options = append(options, "-" + string(name))
		}
	}
	if len(args) == 0 {
		return options
	}
	return fuzzyFind(args[len(args)-1], options)
}
