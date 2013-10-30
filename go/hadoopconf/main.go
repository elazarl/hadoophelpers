package main

import (
	"errors"
	"github.com/GeertJohan/go.linenoise"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elazarl/hadoophelpers/go/lib/hadoopconf"
	"github.com/foize/go.sgr"
	"github.com/jessevdk/go-flags"
	"github.com/elazarl/hadoophelpers/go/lib/table"
	//"github.com/wsxiaoys/terminal"
)

type getOpts struct {}

type setOpts struct {}

type envAddOpts struct {
	Append bool `long:"append" default:"false" description:"append value to environment variable"`
}

type envDelOpts struct {}

type envSetOpts struct {}

type envOpts struct {}

func (o getOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	t := table.New(4)
	c := opt.getConf()
	keys := []string{}
	for _, key := range c.Keys() {
		for _, arg := range args {
			if ok, _ := filepath.Match(arg, key); ok {
				keys = append(keys, key)
				break
			}
		}
	}
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[1].PadLeft = []byte(sgr.FgCyan)
		t.CellConf[2].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[3].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[3].PadRight = []byte(sgr.Reset)
	}
	for _, arg := range keys {
		v, src := c.SourceGet(arg)
		if v == "" && src == hadoopconf.NoSource {
			t.Add("", arg, "", "no property")
		} else {
			t.Add(filepath.Base(src.Source), arg, "=", v)
		}
	}
	fmt.Print(t.String())
	return nil
}

func (o setOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return errors.New("set accepts arguments of the form x=y, no '=' in " + arg)
		}
		opt.getConf().SetIfExist(parts[0], parts[1])
	}
	opt.getConf().Save()
	return nil
}

func assignmentTable() *table.Table {
	t := table.New(4)
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[1].PadLeft = []byte(sgr.FgCyan)
		t.CellConf[2].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[3].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[3].PadRight = []byte(sgr.Reset)
	}
	return t
}

func (o envSetOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.Val)
	v.Val = strings.Join(args[1:], " ")
	t.Add("", "", "now", v.Val)
	if err := opt.getEnv().Save(); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envAddOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.Val)
	v.Prepend(strings.Join(args[1:], " "))
	t.Add("", "", "now", v.Val)
	if err := opt.getEnv().Save(); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envDelOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.Val)
	v.Del(strings.Join(args[1:], " "))
	t.Add("", "", "now", v.Val)
	if err := opt.getEnv().Save(); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	t := assignmentTable()
	c := opt.getEnv()
	keys := []string{}
	for _, key := range c.Keys() {
		for _, arg := range args {
			if ok, _ := filepath.Match(arg, key); ok {
				keys = append(keys, key)
				break
			}
		}
	}
	for _, arg := range keys {
		v := c.Get(arg)
		if v == nil {
			t.Add("", arg, "", "no property")
		} else {
			t.Add(filepath.Base(v.Source), arg, "=", v.Val)
		}
	}
	fmt.Print(t.String())
	return nil
}

func (o *gOpts) UseColors() bool {
	if o.Color == "auto" {
		return IsTerminal(os.Stdout.Fd())
	}
	return o.Color == "true" || o.Color == "t" || o.Color == "1"
}

type gOpts struct {
	Get getOpts `command:"get"`
	Set setOpts `command:"set"`
	SetEnv envSetOpts `command:"envset"`
	AddEnv envAddOpts `command:"envadd"`
	DelEnv envDelOpts `command:"envdel"`
	Env envOpts `command:"env"`
	Verbose bool `short:"v" long:"verbose" default:"true" description:"Show verbose debug information"`
	Color string `long:"color" description:"use colors on output" default:"auto"`
	ConfPath string `short:"c" long:"conf" description:"Set hadoop configuration dir"`
	conf *hadoopconf.HadoopConf
	env hadoopconf.Envs
	executed bool
}

func (opt *gOpts) setConfPath() {
	var p = "."
	if opt.ConfPath != "" {
		p = opt.ConfPath
	} else if os.Getenv("HADOOP_CONF") != "" {
		p = os.Getenv("HADOOP_CONF")
	}
	opt.ConfPath = p
}

func (opt *gOpts) getEnv() hadoopconf.Envs {
	if opt.env != nil {
		return opt.env
	}
	var err error
	opt.setConfPath()
	opt.env, err = hadoopconf.NewEnv(opt.ConfPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return opt.env
}

func (opt *gOpts) getConf() *hadoopconf.HadoopConf {
	if opt.conf != nil {
		return opt.conf
	}
	var err error
	var p = "."
	if opt.ConfPath != "" {
		p = opt.ConfPath
	} else if os.Getenv("HADOOP_CONF") != "" {
		p = os.Getenv("HADOOP_CONF")
	}
	opt.conf, err = hadoopconf.New(p)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return opt.conf
}

var opt gOpts
var conf *hadoopconf.HadoopConf

// bash-like command line parser, splits string to arguments
func parseCommandLine(line string) []string {
	args := []string{}
	type State int
	const (
		REGULAR State = iota
		IN_DQUOTE
		IN_QUOTE
		IN_ESCAPE
	)
	state := REGULAR
	escapePrevState := REGULAR
	next := []byte{}
	seenQuote := false
	for _, r := range []byte(line) {
		switch state {
		case REGULAR:
			if r == ' ' || r == '\t' {
				if len(next) > 0 || seenQuote {
					args = append(args, string(next))
				}
				next = nil
				seenQuote = false
			} else if r == '\'' {
				seenQuote = true
				state = IN_QUOTE
			} else if r == '"' {
				seenQuote = true
				state = IN_DQUOTE
			} else if r == '\\' {
				state = IN_ESCAPE
				escapePrevState = REGULAR
			} else {
				next = append(next, r)
			}
		case IN_QUOTE:
			if r == '\'' {
				state = REGULAR
			} else if r == '\\' {
				state = IN_ESCAPE
				escapePrevState = IN_QUOTE
			} else {
				next = append(next, r)
			}
		case IN_DQUOTE:
			if r == '"' {
				state = REGULAR
			} else if r == '\\' {
				state = IN_ESCAPE
				escapePrevState = IN_DQUOTE
			} else {
				next = append(next, r)
			}
		case IN_ESCAPE:
			next = append(next, r)
			state = escapePrevState
		}
	}
	if len(next) > 0 || seenQuote {
		args = append(args, string(next))
	}
	return args
}

func main() {
	parser := flags.NewParser(&opt, flags.HelpFlag + flags.PassDoubleDash)
	if _, err := parser.ParseArgs(os.Args[1:]); err != nil && opt.executed {
		fmt.Println("dead:", err)
		os.Exit(1)
	}
	opt.getConf() // make sure we have correct conf
	if !opt.executed {
		if !IsTerminal(os.Stdout.Fd()) {
			fmt.Println("terminal not recognized or not supported (windows)")
			return
		}
		completionparser := flags.NewParser(&opt, flags.HelpFlag + flags.PassDoubleDash)
		linenoise.SetCompletionHandler(func (line string) []string {
			args := parseCommandLine(line)
			return Complete(completionparser, args)
		})
		for {
			str, err := linenoise.Line("hadoopconf> ")
			linenoise.AddHistory(str)
			if err != nil {
				if err != linenoise.KillSignalError {
					fmt.Println("Unexpected error: %s", err)
				}
				break
			}
			args := parseCommandLine(str)
			if args, err := parser.ParseArgs(args); err != nil {
				fmt.Println("error:", err)
			} else if len(args) > 0 {
				fmt.Println("excessive arguments:", args)
			}
		}
	}
}
