package main

import "fmt"

type Appendage struct {
	appendageName string
	size          int
}

type Thing struct {
	name string
	app  *Appendage
}

func main() {
	t := &Thing{"charles", nil}
	fmt.Printf("created thing: %v\n", t)

	a := &Appendage{"hand", 3}
	fmt.Printf("created appendage %v\n", a)
	fmt.Printf("assigning an appendage\n")
	t.app = a
	fmt.Printf("assigned appendage. thing: %v\n", t)
}
