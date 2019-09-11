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

			currentVersion := submatches[1]

			if currentVersion == "" {
				panic(errors.New("Unable to find current version in pattern"))
			}

			fmt.Printf("Current version: %s", currentVersion)
			tags := gitRemoteTags(remoteURL)

			if rif.TagPrefix != "" {
				fmt.Printf("Filtering to tags with prefix %s and removing it", rif.TagPrefix)
				tags = filterAndRemovePrefixes(tags, rif.TagPrefix)
				fmt.Printf("Remaining tags: %v\n", tags)
			}

			// TODO assume semver, option to not
			// if semver then sort first
			// (if not semver then will just get the last tag)
			latestVersion := tags[len(tags)-1]
			fmt.Printf("Latest version: %s", latestVersion)

			currentDependencies[remoteURL] = &schema.ManifestDependency{
				Constraint: currentVersion,
				Dependency: &schema.Dependency{
					Source: "git",
					Repo:   remoteURL,
				},
			}

			if latestVersion != currentVersion {
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

func filterAndRemovePrefixes(tags []string, prefix string) []string {
	filtered := []string{}
	for _, s := range tags {
		if strings.HasPrefix(s, prefix) {
			filtered = append(filtered, s[len(prefix):])
		}
	}
	return filtered
}
