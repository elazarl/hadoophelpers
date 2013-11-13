package readline

import (
	"os"
	"path/filepath"
	"strings"
)


func FileCompletions(path string) []string {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() && !strings.HasSuffix(path, "/") {
		return []string{path + "/"}
	}
	if !strings.HasSuffix(path, "/") {
		path = filepath.Dir(path)
	}
	fd, err := os.Open(path)
	if err != nil {
		return []string{}
	}
	dirs, err := fd.Readdirnames(0)
	if err != nil {
		return []string{}
	}
	rv := []string{}
	for _, d := range dirs {
		rv = append(rv, filepath.Join(path, d))
	}
	return rv
}
