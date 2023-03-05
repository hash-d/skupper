// Install Hello World in a 1-1 topology; front-end on pub,
// backend on prv.  Add a new skupper node on a third
// namespace and move part of hello world there.  Once
// good, remove the same from the original namespace (app
// and Skupper).  Validate all good, and move back.
//
// repeat it a few times (or 90% of the alloted test time)
//
// Options:
//
// - remove service first
// - remove link first
// - skupper delete, direct
//
// By default, use a different one each time, but allow
// for selecting a single one
//
// Change everything below this
package main

import "fmt"

func main() {
	fmt.Println("vim-go")
}
