package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ApexCorse/vera"
	"github.com/ApexCorse/vera/codegen"
)

func main() {
	dbcFilePath := flag.String("f", "config.dbc", "DBC file relative path")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("fatal: need build path")
		os.Exit(1)
	}
	buildPath := args[0]
	sourceFilePath := buildPath + "/vera.c"
	headerFilePath := buildPath + "/vera.h"

	dbcFile, err := os.Open(*dbcFilePath)
	if err != nil {
		fmt.Println("fatal: error in opening dbc file: ", err.Error())
		os.Exit(1)
	}

	config, err := vera.Parse(dbcFile)
	if err != nil {
		fmt.Println("fatal: error in parsing dbc file: ", err.Error())
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		fmt.Println("fatal: error in validating dbc file: ", err.Error())
		os.Exit(1)
	}

	sourceFile, err := os.Create(sourceFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating source file: ", err.Error())
		os.Exit(1)
	}
	headerFile, err := os.Create(headerFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating header file: ", err.Error())
		os.Exit(1)
	}

	if err = codegen.GenerateHeader(headerFile, config); err != nil {
		fmt.Println("fatal: error in writing header file: ", err.Error())
		os.Exit(1)
	}
	if err = codegen.GenerateSource(sourceFile, config, "vera.h"); err != nil {
		fmt.Println("fatal: error in writing source file: ", err.Error())
		os.Exit(1)
	}
}
