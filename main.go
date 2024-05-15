package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lesomnus/tiny-short/cmd"
)

func main() {
	c, err := cmd.ParseArgs(os.Args)
	if err != nil {
		panic(err)
	}

	if err := cmd.Run(context.Background(), c); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
