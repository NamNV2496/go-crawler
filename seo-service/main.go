package main

import (
	"github.com/namnv2496/seo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
