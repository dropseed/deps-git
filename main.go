package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"github.com/dropseed/deps/pkg/schema"
)

func main() {
	// --collect or --act will be 1 with current setup...
	inputPath := os.Args[2]
	outputPath := os.Args[3]

	doCollect := flag.Bool("collect", false, "run collect")
	doAct := flag.Bool("act", false, "run act")
	flag.Parse()

	var output *schema.Dependencies

	if *doCollect {
		output = collect(inputPath, outputPath)
	} else if *doAct {
		output = act(inputPath, outputPath)
	} else {
		println("Use --collect or --act")
		os.Exit(1)
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(outputPath, outputBytes, 0644); err != nil {
		panic(err)
	}

}
