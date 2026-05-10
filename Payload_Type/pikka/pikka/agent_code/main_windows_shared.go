//go:build windows && shared
// +build windows,shared

package main

import "C"

//export main
func main() {
	run()
}
