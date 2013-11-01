#include <stdio.h>
#include "readline/readline.h"

// exported in readline.go
extern char** completer(char*, int, int);

void setup_readline_completion() {
	rl_attempted_completion_function = (rl_completion_func_t*)completer;
}
