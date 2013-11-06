package table_test

import (
	"testing"

	"github.com/elazarl/hadoophelpers/go/lib/table"
	. "github.com/robertkrimen/terst"
)

func TestSimpleTable(t *testing.T) {
	Terst(t)
	tbl := table.New(3).
		Add("a", "b", "c").
		Add("1", "2", "3").
		Add("11", "22", "33").
		Add("11", "22", "333")
	Is("\n" + tbl.String(),`
a  b  c
1  2  3
11 22 33
11 22 333
`)
	tbl = table.New(3).Add("a", "b", "c")
	Is(tbl.String(), "a b c\n")
	tbl.CellConf[0].PadRight = []byte{}
	Is(tbl.String(), "ab c\n")
	tbl.CellConf[0].PadRight = []byte("   ")
	Is(tbl.String(), "a   b c\n")
}

