package main

import (
	"fmt"

	"github.com/spektroskop/muon/nd"
)

func main() {
	a := nd.New("a")
	b := nd.New("b")
	c := nd.New("c")

	a.Prev().Link(b)
	a.Prev().Link(c)

	for node := range a.Each() {
		fmt.Println(node.Value)
	}

	fmt.Println()

	b.Unlink()

	for node := range c.Each() {
		fmt.Println(node.Value)
	}
}
