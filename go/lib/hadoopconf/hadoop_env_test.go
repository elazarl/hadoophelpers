package hadoopconf

import (
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestHadoopEnv(t *testing.T) {
	Terst(t)
	defer restoreConf()
	_, err := NewEnv(hadoop1)
	Is(err, nil)
}
