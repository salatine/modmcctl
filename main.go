package main

import (
	"fmt"
	"os"
	"github.com/salatine/modmcctl/internal/cli"
	"github.com/salatine/modmcctl/internal/core"
)

func main() {
	config := cli.ParseConfig()
	if err := core.Run(config); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}
}
