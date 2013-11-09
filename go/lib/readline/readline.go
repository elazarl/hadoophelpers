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
 extern int suppress_enter_key;
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
	replacement, options := Completer(C.GoString(C.rl_line_buffer), start, end)
	raw := C.calloc((C.size_t)(unsafe.Sizeof((*C.char)(nil))), (C.size_t)(len(options) + 2))

	rv := (*[1<<31](*C.char))(raw)
	rv[0] = C.CString(replacement)
	for i, w := range options {
		rv[i+1] = C.CString(w)
	}
	return (**C.char)(raw)
}

func SuppressAppend() {
	C.rl_completion_suppress_append = 1
}

func SuppressEnterKey() {
	C.suppress_enter_key = 1
}

func Readline(prompt string) (string, bool) {
	line := C.readline(C.CString(prompt))
	if line == nil {
		return "", false
	}
	return C.GoString(line), true
}

// DestroyReadline should be called before the program exits,
// to keep the terminal usable
func DestroyReadline() {
	C.rl_deprep_terminal()
}
