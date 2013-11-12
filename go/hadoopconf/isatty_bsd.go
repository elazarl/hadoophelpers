// Copyright 2013 The Go Authors. All rights reserved.

// +build darwin

package main

import (
	"syscall"
	"unsafe"
)

func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
