package table

import (
	"bytes"
	"strconv"
)

type Align int

const (
	Left Align = iota
	Right
)

type CellConf struct {
	Align    Align
	PadRight []byte
	PadLeft  []byte
}

type Table struct {
	CellConf []CellConf
	Data     [][]string
}

func (t *Table) Add(cells ...string) *Table {
	if len(cells) != len(t.CellConf) {
		panic("expected " + strconv.Itoa(len(t.CellConf)) + " got " + strconv.Itoa(len(cells)))
	}
	t.Data = append(t.Data, cells)
	return t
}

func New(size int) *Table {
	conf := make([]CellConf, size)
	dflt := CellConf{Left, []byte{' '}, []byte{}}
	for i := 0; i < size; i++ {
		conf[i] = dflt
	}
	conf[size-1].PadRight = []byte{}
	return NewWithConf(conf)
}

func NewWithConf(conf []CellConf) *Table {
	return &Table{conf, nil}
}

func (t *Table) String() string {
	if len(t.Data) == 0 {
		return "\n"
	}
	b := new(bytes.Buffer)
	max := func(p *int, v int) {
		if *p < v {
			*p = v
		}
	}
	lengths := make([]int, len(t.Data[0]))
	for _, v := range t.Data {
		for i, cell := range v {
			max(&lengths[i], len(cell))
		}
	}
	spc := []byte{' '}
	for _, v := range t.Data {
		for i, cell := range v {
			b.Write(t.CellConf[i].PadLeft)
			if t.CellConf[i].Align == Right {
				b.Write(bytes.Repeat(spc, lengths[i]-len(cell)))
			}
			b.WriteString(cell)
			if t.CellConf[i].Align == Left && i < len(v)-1 {
				b.Write(bytes.Repeat(spc, lengths[i]-len(cell)))
			}
			b.Write(t.CellConf[i].PadRight)
		}
		b.WriteString("\n")
	}
	return b.String()
}
