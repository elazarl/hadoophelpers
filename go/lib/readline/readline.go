package readline

/*
 #cgo darwin CFLAGS: -I/opt/local/include
 #cgo darwin LDFLAGS: -L/opt/local/lib
 #cgo LDFLAGS: -lreadline

 #include <stdio.h>
 #include <stdlib.h>
 #include <string.h>
 #include "readline/readline.h"
 #include "readline/history.h"

 extern void setup_readline_completion();
*/
import "C"
import "unsafe"

func init() {
	C.setup_readline_completion()
}

type CompleteFunc func (text string, start, end int) (replacement string, options []string)

var Completer CompleteFunc

//export completer
func completer(line *C.char, start, end int) **C.char {
	if Completer == nil {
		return nil
	}
	replacement, options := Completer(C.GoString(line), start, end)
	raw := C.calloc((C.size_t)(unsafe.Sizeof((*C.char)(nil))), (C.size_t)(len(options) + 2))

	rv := (*[1<<31](*C.char))(raw)
	rv[0] = C.CString(replacement)
	for i, w := range options {
		rv[i+1] = C.CString(w)
	}
	return (**C.char)(raw)
}

func Readline(prompt string) (string, bool) {
	line := C.readline(C.CString(prompt))
	if line == nil {
		return "", false
	}
	return C.GoString(line), true
}
