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
	d := LevenshteinDistance(idol,candidate)
	if strings.HasPrefix(idol, candidate) {
		d -= 1000 // 1,000 is the infininty of the levenstein distance
	} else if strings.Contains(idol, candidate) {
		d -= 100
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

var CompletionLimit = 20

func Complete(parser *flags.Parser, args []string, partial string) []string {
	parser.Options |= flags.IgnoreUnknown
	opt.completeOpts = []string{}
	defer func() {
		opt.completeOpts = []string{}
		opt.executed = false
	}()
	opt.parser = parser
	parser.ParseArgs(args)
	options := []string{}
	if opt.executed {
		options = opt.completeOpts
	} else {
		for _, group := range parser.Groups {
			options = append(options, getGroupOptions(group)...)
		}
	}
	if partial == "" {
		return options
	}
	rv := fuzzyFind(partial, options)
	if len(rv) >= CompletionLimit {
		return rv[:CompletionLimit]
	}
	return rv
}

func getGroupOptions(group *flags.Group) []string {
	options := []string{}
	for name := range group.LongNames {
		options = append(options, "--"+name)
	}
	for name := range group.ShortNames {
		options = append(options, "-"+string(name))
	}
	for name := range group.Commands {
		options = append(options, name)
	}
	return options
}
