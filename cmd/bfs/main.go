package main

import (
	"github.com/ahmwanes/bfs/pkg/block"
	"github.com/ahmwanes/bfs/pkg/fs"
)

func main() {
	name := "disk.img"
	d, err := block.New(name, uint64(2*block.GB), "12345")
	if err != nil {
		panic(err)
	}
	defer d.Close()

	_, err = fs.NewFS(d)
	if err != nil {
		panic(err)
	}

}
