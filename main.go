package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type remote struct {
	Url            string           `json:"url"`
	ReplaceInFiles []*replaceInFile `json:"replace_in_files"`
}

type replaceInFile struct {
	Filename  string `json:"filename"`
	Pattern   string `json:"pattern"`
	TagPrefix string `json:"tag_prefix"`
}

func main() {
	// --collect or --act will be 1...
	inputPath := os.Args[2]
	outputPath := os.Args[3]

	fmt.Printf("Input: %s\nOutput: %s\n", inputPath, outputPath)

	doCollect := flag.Bool("collect", false, "run collect")
	doAct := flag.Bool("act", false, "run act")
	flag.Parse()

	if !*doCollect && !*doAct {
		println("Use --collect or --act")
		os.Exit(1)
	}

	output := collect(inputPath, outputPath)

	outputBytes, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(outputPath, outputBytes, 0644); err != nil {
		panic(err)
	}

}
