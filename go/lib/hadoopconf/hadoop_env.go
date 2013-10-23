package hadoopconf

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// an evironment variable is a line of the form
// export FOO_OPT="-a -b -c" in a bash environment
// file.
type Env struct {
	Path string
	Vars []*Var
}

type Var struct {
	line int
	Name string
	Val  string
}

// AddIfNew adds a toadd, only if there's no tocheck
// already in the value.
// Use case is
//     v.Update("-Xmx=", "-Xmx=1g")
// which updates occurences of -Xmx=... to -Xmx=1g
func (v *Var) Update(update, newval string) {
	if strings.Contains(v.Val, update) {
		re := regexp.MustCompile(regexp.QuoteMeta(update) + `[^ ]*`)
		v.Val = re.ReplaceAllString(v.Val, newval)
	} else {
		v.Prepend(newval)
	}
}

func (v *Var) Prepend(tok string) {
	if v.Val != "" {
		tok += " "
	}
	v.Val = tok + v.Val
}

func (v *Var) Append(tok string) {
	if v.Val != "" {
		v.Val += " "
	}
	v.Val += tok
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

func NewEnv(path string) (Envs, error) {
	files := []string{}
	for _, d := range []string{ path, filepath.Join(path, "etc", "hadoop"), filepath.Join(path, "conf") } {
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

var exportLine = regexp.MustCompile(`^export ([A-Z0-9_]+)="?([^"]*)"?`)
var exportComment = regexp.MustCompile(`^#\s*export ([A-Z0-9_]+)=`)


func NewEnvFromFile(path string) (*Env, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	env := Env{path, nil}
	scanner := bufio.NewScanner(f)
	for i := 0; scanner.Scan(); i++ {
		if matches := exportLine.FindStringSubmatch(scanner.Text()); matches != nil {
			env.Vars = append(env.Vars, &Var{i, matches[1], matches[2]})
		}
		if matches := exportComment.FindStringSubmatch(scanner.Text()); matches != nil {
			env.Vars = append(env.Vars, &Var{i, matches[1], ""})
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

func (env *Env) GetValue(name string) string {
	v := env.Get(name)
	if v == nil {
		return ""
	}
	return v.Name
}