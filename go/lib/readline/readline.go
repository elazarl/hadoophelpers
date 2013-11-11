package readline

/*
 #cgo darwin CFLAGS: -I/opt/local/include
 #cgo darwin LDFLAGS: /opt/local/lib/libreadline.a /opt/local/lib/libncurses.a
 #cgo linux LDFLAGS: -Wl,-Bstatic -lreadline -ltinfo -Wl,-Bdynamic

 #include <stdio.h>
 #include <stdlib.h>
 #include <string.h>
 #include "readline/readline.h"
 #include "readline/history.h"

 extern void setup_readline_completion();
 extern int suppress_enter_key;
*/
import "C"
import "os"
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

var (
	historyFile = ""
	historyFileCstr *C.char
	historyRead = false
)

func SetHistoryFile(f string) {
	historyFile = f
	historyRead = false
}

func Readline(prompt string) (string, bool) {
	if historyFile != "" && !historyRead {
		C.clear_history()
		C.free(unsafe.Pointer(historyFileCstr))
		historyFileCstr = C.CString(historyFile)
		_, staterr := os.Stat(historyFile)
		if rv, err := C.read_history(historyFileCstr); rv != 0 && !os.IsNotExist(staterr) {
			os.Stderr.WriteString("Cannot read history file "+ historyFile +": " + err.Error() + "\n")
		}
		historyRead = true
	}
	line := C.readline(C.CString(prompt))
	if line == nil {
		return "", false
	}
	C.add_history(line)
	if historyFile != "" {
		C.write_history(historyFileCstr)
	}
	return C.GoString(line), true
}

func AddHistory(line string) {
	str := C.CString(line)
	defer C.free(unsafe.Pointer(str))
	C.add_history(str)
}

// DestroyReadline should be called before the program exits,
// to keep the terminal usable
func DestroyReadline() {
	C.rl_deprep_terminal()
	C.free(unsafe.Pointer(historyFileCstr))
}
