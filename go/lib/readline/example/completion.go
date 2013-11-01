package main

import (
	"github.com/elazarl/hadoophelpers/go/lib/readline"
)

func main() {
	readline.Completer = func(text string, start, end int) (string, []string) {
		return "", []string{"abc", "def", "ghi"}
	}
	readline.Readline("press tab, you should see abc def ghi> ")
	readline.Completer = func(text string, start, end int) (string, []string) {
		return "davidka", []string{"helped", "IDF", "once"}
	}
	readline.Readline("now, the word should be replaced with davidka, and you should see 'helped IDF once'> ")
}
