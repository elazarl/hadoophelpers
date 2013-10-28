package main

import (
	"testing"

	. "github.com/robertkrimen/terst"
)

func TestSimpleParseLine(t *testing.T) {
	Terst(t)
	Is(parseCommandLine(`a great success`), []string{"a", "great", "success"})
	Is(parseCommandLine(`a    great    success  `), []string{"a", "great", "success"})
	Is(parseCommandLine(`a`), []string{"a"})
	Is(parseCommandLine(``), []string{})
	Is(parseCommandLine(`    `), []string{})
}

func TestParseLineWithQuotes(t *testing.T) {
	Terst(t)
	Is(parseCommandLine(`a "great" success`), []string{"a", "great", "success"})
	Is(parseCommandLine(`a    gre'a't    success  `), []string{"a", "great", "success"})
	Is(parseCommandLine(`"a"`), []string{"a"})
	Is(parseCommandLine(`""`), []string{""})
	Is(parseCommandLine(`""  `), []string{""})
	Is(parseCommandLine(`  ""  `), []string{""})
	Is(parseCommandLine(`  ""`), []string{""})
	Is(parseCommandLine(`''  ''  `), []string{"", ""})
}

func TestParseLineNestedQuotes(t *testing.T) {
	Terst(t)
	Is(parseCommandLine(`"give me 'that'"`), []string{"give me 'that'"})
	Is(parseCommandLine(`  "give me 'that'"  `), []string{"give me 'that'"})
	Is(parseCommandLine(`the 'man in "spain"'"!" got hain`), []string{"the", "man in \"spain\"!", "got", "hain"})
	Is(parseCommandLine(`"'\""`), []string{`'"`})
	Is(parseCommandLine(`'\\'`), []string{`\`})
}
