package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ApexCorse/vera"
	"github.com/ApexCorse/vera/codegen"
	"github.com/ApexCorse/vera/codegen/autodevkit"
)

func main() {
	dbcFilePath := flag.String("f", "config.dbc", "DBC file relative path")
	sdk := flag.String("sdk", "", "SDK to generate the adapters for")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("fatal: need build path")
		os.Exit(1)
	}
	buildPath := args[0]
	sourceFilePath := buildPath + "/vera.c"
	headerFilePath := buildPath + "/vera.h"

	switch *sdk {
	case "autodevkit":
		// can throw
		autodevkitGeneration(buildPath)
	case "":
		break
	default:
		fmt.Printf("fatal: sdk '%s' not supported\n", *sdk)
		os.Exit(1)
	}

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

func autodevkitGeneration(buildPath string) {
	autodevkitSourceFilePath := buildPath + "/vera_autodevkit.c"
	autodevkitHeaderFilePath := buildPath + "/vera_autodevkit.h"

	autodevkitSourceFile, err := os.Create(autodevkitSourceFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating autodevkit source file: ", err.Error())
		os.Exit(1)
	}

	autodevkitHeaderFile, err := os.Create(autodevkitHeaderFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating autodevkit include file: ", err.Error())
		os.Exit(1)
	}

	if err := autodevkit.GenerateSource(autodevkitSourceFile); err != nil {
		fmt.Println("fatal: error in writing autodevkit source file: ", err.Error())
		os.Exit(1)
	}

	if err := autodevkit.GenerateHeader(autodevkitHeaderFile); err != nil {
		fmt.Println("fatal: error in writing autodevkit source file: ", err.Error())
		os.Exit(1)
	}
}
