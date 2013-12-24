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
	d := LevenshteinDistance(idol, candidate)
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
		opt.completionCandidate = ""
		opt.executed = false
	}()
	opt.completionCandidate = partial
	opt.parser = parser
	parser.ParseArgs(args)
	options := []string{}
	if opt.executed {
		options = opt.completeOpts
	} else {
		for _, command := range parser.Commands() {
			options = append(options, command.Name)
		}
		options = append(options, groupsOptions(parser.Groups())...)
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

func groupsOptions(groups []*flags.Group) []string {
	options := []string{}
	for _, group := range groups {
		options = append(options, groupOptions(group)...)
	}
	return options
}
// groupOptions returns available options for a group as a string.
// For example, for the group from
//     struct {
//         foo string `short:"f" long:"foo"`
//     }
// it'll return []string{"-f", "foo"}
func groupOptions(group *flags.Group) []string {
	options := []string{}
	for _, option := range group.Options() {
		if option.ShortName != 0 {
			options = append(options, "-"+string(option.ShortName))
		}
		if option.LongName != "" {
			options = append(options, "--"+option.LongName)
		}
	}
	return options
}
