package main

import (
	"testing"

	"github.com/jessevdk/go-flags"
	. "github.com/robertkrimen/terst"
)

func TestCompleteOptions(t *testing.T) {
	Terst(t)
	var opt envAddOpts
	parser := flags.NewParser(&opt, flags.HelpFlag+flags.PassDoubleDash+flags.IgnoreUnknown)
	Is(Complete(parser, []string{"-v"}, "--h"), []string{"--help", "-h", "--backup", "--append"})
	Is(Complete(parser, []string{}, "--"), []string{"--help", "--backup", "--append", "-h"})
}
