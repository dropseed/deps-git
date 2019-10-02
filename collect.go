package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/dropseed/deps/pkg/schema"
)

func collect(inputPath, outputPath string) *schema.Dependencies {

	// TODO should use inputPath to join with filename?

	currentDependencies := map[string]*schema.ManifestDependency{}
	updatedDependencies := map[string]*schema.ManifestDependency{}

	for remoteURL, remote := range loadRemotesFromEnv() {
		fmt.Printf("Collecting remote %s\n", remoteURL)

		for _, rif := range remote.ReplaceInFiles {
			regex := rif.regex()

			fileStr := rif.readFile()
			submatches := regex.FindStringSubmatch(fileStr)

			if len(submatches) < 1 {
				panic(fmt.Errorf("Pattern not found in file\n\n  Pattern: %s\n  Filename: %s", rif.Pattern, rif.Filename))
			}

			currentVersion := submatches[1]

			if currentVersion == "" {
				panic(errors.New("Unable to find current version in pattern"))
			}

			fmt.Printf("Current version: %s\n", currentVersion)
			tags := gitRemoteTags(remoteURL)

			latestVersion := rif.getLatestTag(tags)
			fmt.Printf("Latest version: %s\n", latestVersion)

			currentDependencies[remoteURL] = &schema.ManifestDependency{
				Constraint: currentVersion,
				Dependency: &schema.Dependency{
					Source: "git",
					Repo:   remoteURL,
				},
			}

			if latestVersion != "" && latestVersion != currentVersion {
				updatedDependencies[remoteURL] = &schema.ManifestDependency{
					Constraint: latestVersion,
					Dependency: &schema.Dependency{
						Source: "git",
						Repo:   remoteURL,
					},
				}
			}
		}
	}

	output := &schema.Dependencies{
		Manifests: map[string]*schema.Manifest{
			"": &schema.Manifest{
				Current: &schema.ManifestVersion{
					Dependencies: currentDependencies,
				},
				Updated: &schema.ManifestVersion{
					Dependencies: updatedDependencies,
				},
			},
		},
	}

	return output
}

func gitRemoteTags(url string) []string {
	cmd := exec.Command("git", "ls-remote", "--tags", url)
	lsOutput, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	// Preserve the order, but remove duplicates as we go
	tags := []string{}
	tagsSeen := map[string]bool{}

	for _, line := range strings.Split(string(lsOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		tag := fields[1]
		tag = strings.Replace(tag, "refs/tags/", "", 1)

		// Annotated tags
		if strings.HasSuffix(tag, "^{}") {
			tag = tag[:len(tag)-3]
		}

		if _, seen := tagsSeen[tag]; !seen {
			tags = append(tags, tag)
			tagsSeen[tag] = true
		}
	}

	return tags
}
