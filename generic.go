// +build !plan9

package main

func IsMundaneError(err error) bool {
	return false
}
