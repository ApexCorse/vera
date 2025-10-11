package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ApexCorse/vera/internal/parser"
)

func main() {
	dbcFilePath := flag.String("f", "config.dbc", "DBC file relative path")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("fatal: need build path")
		os.Exit(1)
	}

	dbcFile, err := os.Open(*dbcFilePath)
	if err != nil {
		fmt.Println("fatal: error in opening dbc file: ", err.Error())
		os.Exit(1)
	}

	config, err := parser.Parse(dbcFile)
	if err != nil {
		fmt.Println("fatal: error in parsing dbc file: ", err.Error())
		os.Exit(1)
	}

	fmt.Println(config)
}
