package hadoopconf

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// an evironment variable is a line of the form
// export FOO_OPT="-a -b -c" in a bash environment
// file.
type Env struct {
	Path string
	Vars []*Var
}

type Var struct {
	modified bool
	line     int
	Comment  string
	Source   string
	Name     string
	val      string
}

func (v *Var) GetVal() string {
	return v.val
}

func (v *Var) SetVal(val string) {
	v.modified = true
	v.val = val
}

// Update adds a toadd, only if there's no tocheck
// already in the value.
// Use case is
//     v.Update("-Xmx=", "-Xmx=1g")
// which updates occurences of -Xmx=... to -Xmx=1g
func (v *Var) Update(update, newval string) {
	if strings.Contains(v.GetVal(), update) {
		re := regexp.MustCompile(regexp.QuoteMeta(update) + `[^ ]*`)
		v.SetVal(re.ReplaceAllString(v.GetVal(), newval))
	} else {
		v.Prepend(newval)
	}
}

// Del deletes value from the variable
func (v *Var) Del(val string) {
	switch {
	case strings.Contains(v.GetVal(), " "+val+" "):
		v.SetVal(strings.Replace(v.GetVal(), " "+val+" ", " ", -1))
	case strings.HasSuffix(v.GetVal(), " "+val):
		v.SetVal(v.GetVal()[:len(v.GetVal())-len(val)-1])
	case strings.HasPrefix(v.GetVal(), val+" "):
		v.SetVal(v.GetVal()[len(val)+1:])
	case v.GetVal() == val:
		v.SetVal("")
	}
}

func (v *Var) Prepend(tok string) {
	if v.GetVal() != "" {
		tok += " "
	}
	v.SetVal(tok + v.GetVal())
}

func (v *Var) Append(tok string) {
	if v.GetVal() != "" {
		v.SetVal(v.GetVal() + " ")
	}
	v.SetVal(v.GetVal() + tok)
}

type Envs []*Env

func (envs Envs) Get(name string) *Var {
	for _, env := range envs {
		if r := env.Get(name); r != nil {
			return r
		}
	}
	return nil
}

func (envs Envs) Keys() []string {
	keys := []string{}
	for _, env := range envs {
		keys = append(keys, env.Keys()...)
	}
	return keys
}

func (envs Envs) Save(backup bool) error {
	for _, env := range envs {
		if err := env.Save(backup); err != nil {
			return err
		}
	}
	return nil
}

/*snprintf(systembuf, sizeof(systembuf), "echo 'attach %d\nbt\nquit' | gdb -quiet _test_main.out ", getpid());*/

func NewEnv(path string) (Envs, error) {
	files := []string{}
	for _, d := range []string{path, filepath.Join(path, "etc", "hadoop"), filepath.Join(path, "conf")} {
		r, err := filepath.Glob(filepath.Join(d, "*-env.sh"))
		if err == nil && len(r) > 0 {
			files = r
			break
		}
	}
	if len(files) == 0 {
		return nil, errors.New("no *-env.sh files found in " + path)
	}
	envs := Envs{}
	for _, file := range files {
		if env, err := NewEnvFromFile(file); err != nil {
			return nil, err
		} else {
			envs = append(envs, env)
		}
	}
	return envs, nil
}

var exportLine = regexp.MustCompile(`^\s*(#?)\s*export\s+([A-Z0-9_]+)=(.*)$`)

// parseExport is a poor man's parser of bash export line.
// If you don't abuse bash too much - it should work.
func parseExport(filename string, lineno int, line string) *Var {
	if matches := exportLine.FindStringSubmatch(line); matches != nil {
		s := matches[3]
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			s = s[1 : len(s)-1]
		}
		comment := ""
		if matches[1] == "#" {
			comment = s
			s       = ""
		}
		return &Var{false, lineno, comment, filename, matches[2], s}
	}
	return nil
}

func NewEnvFromFile(path string) (*Env, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	env := Env{path, nil}
	scanner := bufio.NewScanner(f)
	for i := 0; scanner.Scan(); i++ {
		if v := parseExport(path, i, scanner.Text()); v != nil {
			env.Vars = append(env.Vars, v)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &env, nil
}

func (env *Env) Get(name string) *Var {
	for _, v := range env.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (env *Env) Keys() []string {
	keys := []string{}
	for _, v := range env.Vars {
		keys = append(keys, v.Name)
	}
	return keys
}

func (env *Env) GetValue(name string) string {
	v := env.Get(name)
	if v == nil {
		return ""
	}
	return v.Name
}

func (env *Env) Save(backup bool) error {
	out, err := ioutil.TempFile("/tmp", "gohadoop")
	if err != nil {
		return err
	}
	f, err := os.Open(env.Path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	varlines := make(map[int]*Var)
	for _, v := range env.Vars {
		if v.modified {
			varlines[v.line] = v
		}
	}
	for i := 0; scanner.Scan(); i++ {
		if v, ok := varlines[i]; ok {
			out.WriteString("export " + v.Name + "=\"" + v.GetVal() + "\"\n")
		} else {
			out.WriteString(scanner.Text() + "\n")
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if backup {
		os.Rename(env.Path, env.Path+time.Now().Format(".2006-01-02_15_04.000"))
	}
	return os.Rename(out.Name(), env.Path)
}
