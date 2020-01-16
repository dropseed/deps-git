package main

import (
	"fmt"
	"strings"

	"github.com/dropseed/deps/pkg/schema"
)

func act(inputPath, outputPath string) *schema.Dependencies {

	remotes := loadRemotesFromEnv()

	dependencies, err := schema.NewDependenciesFromJSONPath(inputPath)
	if err != nil {
		panic(err)
	}

	for _, manifest := range dependencies.Manifests {
		if manifest.Updated != nil {
			for name, updatedDep := range manifest.Updated.Dependencies {
				updatedVersion := updatedDep.Constraint

				remote := remotes[name]
				for _, rif := range remote.ReplaceInFiles {
					regex := rif.regex()
					fileStr := rif.readFile()

					submatch := regex.FindStringSubmatch(fileStr)
					currentStr := submatch[0]
					foundVersion := submatch[1] // the single expected capture group

					replacement := strings.Replace(currentStr, foundVersion, updatedVersion, 1)
					fmt.Printf("Replacing %s with %s in %s\n", currentStr, replacement, rif.Filename)

					result := regex.ReplaceAllString(fileStr, replacement)
					rif.writeFile(result)
				}
			}
		}
	}

	return dependencies
}
