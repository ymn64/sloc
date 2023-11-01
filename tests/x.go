//go:build ignore

package main

import "fmt"

func main() {
	/* Multiline
	comment */fmt.Println("Hello world")

	// This is supposed to be 4 sloc
	fmt.Println(`/*
		Fake
		comment
	*/`)
}
