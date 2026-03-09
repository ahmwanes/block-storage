package main

import (
	"fmt"

	"github.com/ahmwanes/bfs/pkg/block"
)

func main() {
	name := "disk.img"
	d, err := block.New(name, uint64(2*block.GB), "12345")
	if err != nil {
		panic(err)
	}

	w := []byte("Hello World!")

	d.WriteBlock(0, w)
	d.WriteBlock(100, w)

	r, err := d.ReadBlock(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r))

}
