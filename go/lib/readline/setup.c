#include <stdio.h>
#include "readline/readline.h"

// exported in readline.go
extern char** completer(char*, int, int);

void enter_hook() {
}

int suppress_enter_key = 0;

int enter(int state, int key) {
	if (suppress_enter_key) {
		suppress_enter_key = 0;
		return 0;
	} else {
		return rl_newline(state, key);
	}
}

void setup_readline_completion() {
	rl_attempted_completion_function = (rl_completion_func_t*)completer;
	rl_sort_completion_matches = 0;
	rl_bind_key(RETURN, enter);
}
