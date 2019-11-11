// +build plan9

package main

import (
	"os"
	"syscall"
)

func IsMundaneError(err error) bool {
	if v, ok := err.(*os.PathError); ok {
		if v, ok := v.Err.(syscall.ErrorString); ok && v == syscall.EINTR {
			return true
		}
	}
	return false
}
