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

type envAddOpts struct {}

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

func (o envSetOpts) Execute(args []string) error {
	opt.executed = true
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := table.New(3)
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgCyan)
		t.CellConf[1].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[2].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[2].PadRight = []byte(sgr.Reset)
	}
	t.Add(v.Name, "was", v.Val)
	fmt.Println(v.Name, "was", v.Val)
	v.Val = strings.Join(args[1:], " ")
	t.Add("", "now", v.Val)
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
	t := table.New(3)
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgCyan)
		t.CellConf[1].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[2].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[2].PadRight = []byte(sgr.Reset)
	}
	t.Add(v.Name, "was", v.Val)
	fmt.Println(v.Name, "was", v.Val)
	v.Append(strings.Join(args[1:], " "))
	t.Add("", "now", v.Val)
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
	t := table.New(3)
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
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgCyan)
		t.CellConf[1].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[2].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[2].PadRight = []byte(sgr.Reset)
	}
	for _, arg := range keys {
		v := c.Get(arg)
		if v == nil {
			t.Add(arg, "", "no property")
		} else {
			t.Add(arg, "=", v.Val)
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
		for {
			str, err := linenoise.Line("hadoopconf> ")
			linenoise.AddHistory(str)
			if err != nil {
				if err != linenoise.KillSignalError {
					fmt.Println("Unexpected error: %s", err)
				}
				break
			}
			if args, err := parser.ParseArgs(strings.Fields(str)); err != nil {
				fmt.Println("error:", err)
			} else if len(args) > 0 {
				fmt.Println("excessive arguments:", args)
			}
		}
	}
}
