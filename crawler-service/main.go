package main

import (
	"github.com/namnv2496/crawler/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
