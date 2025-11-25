package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ApexCorse/vera"
	"github.com/ApexCorse/vera/codegen"
	"github.com/ApexCorse/vera/codegen/autodevkit"
	"github.com/ApexCorse/vera/codegen/stm32hal"
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

	dbcFile, err := os.Open(*dbcFilePath)
	if err != nil {
		fmt.Println("fatal: error in opening dbc file: ", err.Error())
		os.Exit(1)
	}
	defer dbcFile.Close()

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
	defer sourceFile.Close()
	headerFile, err := os.Create(headerFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating header file: ", err.Error())
		os.Exit(1)
	}
	defer headerFile.Close()

	if err = codegen.GenerateHeader(headerFile, config); err != nil {
		fmt.Println("fatal: error in writing header file: ", err.Error())
		os.Exit(1)
	}
	if err = codegen.GenerateSource(sourceFile, config, "vera.h"); err != nil {
		fmt.Println("fatal: error in writing source file: ", err.Error())
		os.Exit(1)
	}

	switch *sdk {
	case "autodevkit":
		// can throw
		autodevkitGeneration(buildPath, config)
	case "stm32hal":
		stm32halGeneration(buildPath, config)
	case "":
	default:
		fmt.Printf("fatal: sdk '%s' not supported\n", *sdk)
		os.Exit(1)
	}
}

func autodevkitGeneration(buildPath string, config *vera.Config) {
	autodevkitSourceFilePath := buildPath + "/vera_autodevkit.c"
	autodevkitHeaderFilePath := buildPath + "/vera_autodevkit.h"

	autodevkitSourceFile, err := os.Create(autodevkitSourceFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating autodevkit source file: ", err.Error())
		os.Exit(1)
	}
	defer autodevkitSourceFile.Close()

	autodevkitHeaderFile, err := os.Create(autodevkitHeaderFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating autodevkit include file: ", err.Error())
		os.Exit(1)
	}
	defer autodevkitHeaderFile.Close()

	if err := autodevkit.GenerateSource(autodevkitSourceFile, config); err != nil {
		fmt.Println("fatal: error in writing autodevkit source file: ", err.Error())
		os.Exit(1)
	}

	if err := autodevkit.GenerateHeader(autodevkitHeaderFile, config); err != nil {
		fmt.Println("fatal: error in writing autodevkit include file: ", err.Error())
		os.Exit(1)
	}
}

func stm32halGeneration(buildPath string, config *vera.Config) {
	stm32halSourceFilePath := buildPath + "/vera_stm32hal.c"
	stm32halHeaderFilePath := buildPath + "/vera_stm32hal.h"

	stm32halSourceFile, err := os.Create(stm32halSourceFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating stm32hal source file: ", err.Error())
		os.Exit(1)
	}
	defer stm32halSourceFile.Close()

	stm32halHeaderFile, err := os.Create(stm32halHeaderFilePath)
	if err != nil {
		fmt.Println("fatal: error in creating stm32hal include file: ", err.Error())
		os.Exit(1)
	}
	defer stm32halHeaderFile.Close()

	if err := stm32hal.GenerateSource(stm32halSourceFile, config); err != nil {
		fmt.Println("fatal: error in writing stm32hal source file: ", err.Error())
		os.Exit(1)
	}

	if err := stm32hal.GenerateHeader(stm32halHeaderFile, config); err != nil {
		fmt.Println("fatal: error in writing stm32hal include file: ", err.Error())
		os.Exit(1)
	}
}
