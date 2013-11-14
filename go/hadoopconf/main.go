package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/elazarl/hadoophelpers/go/lib/hadoopconf"
	"github.com/elazarl/hadoophelpers/go/lib/readline"
	"github.com/elazarl/hadoophelpers/go/lib/table"
	"github.com/foize/go.sgr"
	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(&opt, flags.HelpFlag|flags.PassDoubleDash|flags.IgnoreUnknown)
	if _, err := parser.ParseArgs(os.Args[1:]); err != nil && opt.executed {
		fmt.Println("dead:", err)
		os.Exit(1)
	}
	if opt.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	if !opt.executed {
		defer readline.DestroyReadline()
		u, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot get current user:", err)
			return
		}
		readline.SetHistoryFile(filepath.Join(u.HomeDir, ".hadoopconf_paths"))
		opt.interactive = true
		// make sure we ask for configuration
		opt.getConf()
		readline.SetHistoryFile(filepath.Join(u.HomeDir, ".hadoopconf_history"))
		if !IsTerminal(os.Stdout.Fd()) {
			fmt.Println("terminal not recognized or not supported (windows)")
			return
		}
		readline.Completer = func(line string, start, end int) (string, []string) {
			completionparser := flags.NewParser(&opt, flags.HelpFlag|flags.PassDoubleDash|flags.IgnoreUnknown)
			opt.executed = false
			args := parseCommandLine(line[:end])
			if len(line) == 0 || line[end-1] == ' ' || line[end-1] == '\t' {
				return "", Complete(completionparser, args, "")
			}
			return "", Complete(completionparser, args[:len(args)-1], args[len(args)-1])
		}
		for {
			str, ok := readline.Readline("hadoopconf> ")
			if !ok {
				break
			}
			opt.completeOpts = nil
			args := parseCommandLine(str)
			if args, err := parser.ParseArgs(args); err != nil {
				fmt.Println("error:", err)
			} else if len(args) > 0 {
				fmt.Println("excessive arguments:", args)
			}
		}
	}
}

type getOpts struct{}

type setOpts struct {
	Backup bool `long:"backup" default:"true" description:"save backup of modified files in the form of oldfile.timestamp"`
}

type envAddOpts struct {
	Append bool `long:"append" default:"false" description:"append value to environment variable"`
	Backup bool `long:"backup" default:"true" description:"save backup of modified files in the form of oldfile.timestamp"`
}

type envDelOpts struct {
	Backup bool `long:"backup" default:"true" description:"save backup of modified files in the form of oldfile.timestamp"`
}

type envSetOpts struct{
	Backup bool `long:"backup" default:"true" description:"save backup of modified files in the form of oldfile.timestamp"`
}

type envOpts struct{}

type statOpts struct{}

func (o getOpts) Execute(args []string) error {
	opt.executed = true
	if opt.completeOpts != nil {
		groups := getmygroups(o, &opt)
		options := getGroupOptions(groups)
		opt.completeOpts = append(options, opt.getConf().Keys()...)
		return nil
	}
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
	if opt.completeOpts != nil {
		options := getGroupOptions(getmygroups(o, &opt))
		if strings.HasSuffix(opt.completionCandidate, "=") {
			s, src := opt.getConf().SourceGet(opt.completionCandidate[:len(opt.completionCandidate)-1])
			if src != hadoopconf.NoSource {
				opt.completeOpts = []string{s}
			}
		} else {
			for _, v := range opt.getConf().Keys() {
				opt.completeOpts = append(opt.completeOpts, v+"=")
			}
			for _, v := range options {
				opt.completeOpts = append(opt.completeOpts, v+" ")
			}
			readline.SuppressAppend()
			readline.SuppressEnterKey()
		}
		return nil
	}
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
	opt.getConf().Save(o.Backup)
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
	if opt.completeOpts != nil {
		options := getGroupOptions(getmygroups(o, &opt))
		if len(args) == 0 {
			opt.completeOpts = append(options, opt.getEnv().Keys()...)
			readline.SuppressEnterKey()
		} else {
			if v := opt.getEnv().Get(args[0]); v != nil {
				if v.GetVal() != "" {
					opt.completeOpts = append(opt.completeOpts, v.GetVal())
				}
				if v.Comment != "" {
					opt.completeOpts = append(opt.completeOpts, v.Comment)
				}
				if strings.HasSuffix(args[0], "_HOME") || strings.HasSuffix(args[0], "_DIR") {
					opt.completeOpts = append(opt.completeOpts, readline.FileCompletions(opt.completionCandidate)...)
					readline.SuppressAppend()
				}
			}
		}
		return nil
	}
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.GetVal())
	v.SetVal(strings.Join(args[1:], " "))
	t.Add("", "", "now", v.GetVal())
	if err := opt.getEnv().Save(o.Backup); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envAddOpts) Execute(args []string) error {
	opt.executed = true
	if opt.completeOpts != nil {
		options := getGroupOptions(getmygroups(o, &opt))
		if len(args) == 0 {
			opt.completeOpts = append(options, opt.getEnv().Keys()...)
		}
		return nil
	}
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.GetVal())
	v.Prepend(strings.Join(args[1:], " "))
	t.Add("", "", "now", v.GetVal())
	if err := opt.getEnv().Save(o.Backup); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envDelOpts) Execute(args []string) error {
	opt.executed = true
	if opt.completeOpts != nil {
		options := getGroupOptions(getmygroups(o, &opt))
		if len(args) == 0 {
			opt.completeOpts = append(options, opt.getEnv().Keys()...)
		} else {
			if v := opt.getEnv().Get(args[0]); v != nil {
				opt.completeOpts = append(opt.completeOpts, parseCommandLine(v.GetVal())...)
			}
		}
		return nil
	}
	if len(args) == 0 {
		return errors.New("get must have nonzero number arguments")
	}
	v := opt.getEnv().Get(args[0])
	if v == nil {
		fmt.Println("No such variable", v)
	}
	t := assignmentTable()
	t.Add(filepath.Base(v.Source), v.Name, "was", v.GetVal())
	v.Del(strings.Join(args[1:], " "))
	t.Add("", "", "now", v.GetVal())
	if err := opt.getEnv().Save(o.Backup); err != nil {
		return err
	}
	fmt.Print(t.String())
	return nil
}

func (o envOpts) Execute(args []string) error {
	opt.executed = true
	if opt.completeOpts != nil {
		options := getGroupOptions(getmygroups(o, &opt))
		opt.completeOpts = append(options, opt.getEnv().Keys()...)
		return nil
	}
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
			t.Add(filepath.Base(v.Source), arg, "=", v.GetVal())
		}
	}
	fmt.Print(t.String())
	return nil
}

func (stat *statOpts) Execute(args []string) error {
	t := table.New(2)
	c := opt.getConf()
	t.Add("core-site.xml", sgr.FgYellow+c.CoreSite.Conf.Source())
	t.Add("hdfs-site.xml", sgr.FgYellow+c.HdfsSite.Conf.Source())
	t.Add("mapred-site.xml", sgr.FgYellow+c.MapredSite.Conf.Source())
	if c.YarnSite.Default != nil {
		t.Add("yarn-site.xml", sgr.FgYellow+c.YarnSite.Conf.Source())
	}
	t.Add("core-default.xml", c.CoreSite.Default.Source())
	t.Add("hdfs-default.xml", c.HdfsSite.Default.Source())
	t.Add("mapred-default.xml", c.MapredSite.Default.Source())
	// This is not mistake, the yarn-site.xml may not exist now, the default must exist
	if c.YarnSite.Default != nil {
		t.Add("yarn-default.xml", c.YarnSite.Default.Source())
	}
	for _, env := range opt.getEnv() {
		t.Add(filepath.Base(env.Path), sgr.FgGreen+env.Path)
	}
	if opt.UseColors() {
		t.CellConf[0].PadLeft = []byte(sgr.FgGrey)
		t.CellConf[0].PadRight = []byte(" " + sgr.ResetForegroundColor)
		t.CellConf[1].PadLeft = []byte(sgr.ResetForegroundColor + sgr.Bold)
		t.CellConf[1].PadRight = []byte(sgr.Reset)
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

type helpOpts struct{}

func (helpOpts) Execute(args []string) error {
	opt.Help = true
	return nil
}

type gOpts struct {
	Get      getOpts    `command:"get"`
	Set      setOpts    `command:"set"`
	SetEnv   envSetOpts `command:"envset"`
	AddEnv   envAddOpts `command:"envadd"`
	DelEnv   envDelOpts `command:"envdel"`
	Stat     statOpts   `command:"stat"`
	Env      envOpts    `command:"env"`
	HelpCmd  helpOpts   `command:"help"`
	Help     bool       `short:"h" long:"help" default:"false" description:"print help"`
	Verbose  bool       `short:"v" long:"verbose" default:"false" description:"Show verbose debug information"`
	Color    string     `long:"color" description:"use colors on output" default:"auto"`
	ConfPath string     `short:"c" long:"conf" description:"Set hadoop configuration dir"`
	JarsPath string     `short:"j" long:"jars" description:"where hadoop's jar are (also searches in DIR/share/hadoop/...), = conf dir if empty"`
	conf     *hadoopconf.HadoopConf
	env      hadoopconf.Envs
	executed bool
	// set this to []string{} if you want command line options to autocomplete instead of executing themselves
	completeOpts        []string
	completionCandidate string
	parser              *flags.Parser
	// marks whether or not we're in interactive mode
	interactive bool
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
	var jarsPath = opt.JarsPath
	if jarsPath == "" {
		jarsPath = p
	}
	jars, err := hadoopconf.Jars(jarsPath)
	if err != nil {
		fmt.Println("cannot find hadoop jars. Specify explicitly with -j/--jars")
		if opt.interactive {
			var ok bool
			if opt.JarsPath, ok = readline.Readline("enter path for hadoop's jars [tab complete files]: "); ok {
				return opt.getConf()
			}
		}
		if opt.Verbose {
			fmt.Print(err)
		}
		os.Exit(1)
	}
	opt.conf, err = hadoopconf.New(p, jars)
	if err != nil {
		fmt.Println("cannot find hadoop configuration. Specify explicitly with -c/--conf")
		// try to guess hadoop location from popular locations
		possibleConfs := map[string]*hadoopconf.HadoopConf{}
		for _, l := range []string{"/etc/hadoop", "/etc/hadoop/*", "/var/run/cloudera-scm-agent/process/*"} {
			if opt.Verbose {
				fmt.Println("Checking", l)
			}
			paths, err := filepath.Glob(l)
			if err != nil {
				if opt.Verbose {
					fmt.Println(err)
				}
				break
			}
			for _, p := range paths {
				if opt.Verbose {
					fmt.Println("Scanning", p)
				}
				if _, err := os.Stat(p); os.IsNotExist(err) {
					continue
				}
				conf, err := hadoopconf.New(p, jars)
				if err != nil {
					if opt.Verbose {
						fmt.Println(err)
					}
					continue
				}
				base := filepath.Dir(conf.CoreSite.Conf.Source())
				possibleConfs[base] = conf
			}
		}
		if opt.interactive {
			var ok bool
			m := map[int]string{}
			i := 0
			prompt := "enter path for hadoop's configuration files [tab complete files]: "
			if len(possibleConfs) > 0 {
				fmt.Println("Automatically recognized existing hadoop configuration:")
				prompt = "Enter path for hadoop's configuration files, or a number from paths above: "
			}
			for k := range possibleConfs {
				m[i] = k
				if opt.UseColors() {
					fmt.Print(sgr.FgRed, i, sgr.Reset, "] ", sgr.ResetBackgroundColor, sgr.Bold, sgr.FgGreen, k, "\n", sgr.Reset)
				} else {
					fmt.Print(i, "] ", k, "\n")
				}
				i++
			}
			if opt.ConfPath, ok = readline.Readline(prompt); ok {
				if i, err := strconv.Atoi(opt.ConfPath); err == nil && i < len(m) && i >= 0 {
					opt.ConfPath = m[i]
				}
				return opt.getConf()
			}
		} else if opt.Verbose {
			fmt.Print(err)
		}
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

// given
// type T struct {
//     foo Foo `command:"moo"`
//     bar Bar `command:"maa"`
// }
// getField(Foo{}, T{}) == "moo"
// getField(Bar{}, T{}) == "maa"
// getField(Baz{}, T{}) panics
func getField(typ interface{}, strct interface{}) string {
	v := reflect.TypeOf(strct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Type.AssignableTo(reflect.TypeOf(typ)) {
			return reflect.StructTag(f.Tag).Get("command")
		}
	}
	panic("cannot find type in struct")
}

func getmygroups(o, strct interface{}) *flags.Group {
	field := getField(o, strct)
	for _, group := range opt.parser.Groups {
		if mygroup, ok := group.Commands[field]; ok {
			return mygroup
		}
	}
	panic("my field not avail")
}
