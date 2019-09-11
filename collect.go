package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func collect(inputPath, outputPath string) map[string]interface{} {
	currentDependencies := map[string]interface{}{}
	updatedDependencies := map[string]interface{}{}

	if remotesStr := os.Getenv("DEPS_SETTING_REMOTES"); remotesStr != "" {
		var remotes []remote
		if err := json.Unmarshal([]byte(remotesStr), &remotes); err != nil {
			println("Invalid remotes")
			os.Exit(1)
		}

		for _, remote := range remotes {
			// println(remote)
			println(remote.Url)

			// tag_prefix
			for _, rif := range remote.ReplaceInFiles {
				regex := regexp.MustCompile(rif.Pattern)

				// TODO use inputPath as base?
				fileBytes, err := ioutil.ReadFile(rif.Filename)
				if err != nil {
					panic(err)
				}
				fileStr := string(fileBytes)
				submatches := regex.FindAllStringSubmatch(fileStr, -1)

				// currentStr := submatches[0][0]
				currentVersion := submatches[0][1]

				if currentVersion == "" {
					panic(errors.New("Unable to find current version in pattern"))
				}

				fmt.Printf("Current version: %s", currentVersion)

				// replacement := currentStr
				// println(replacement)

				// get available version
				tags := gitRemoteTags(remote.Url)
				fmt.Printf("%+v", tags)

				if rif.TagPrefix != "" {
					tags = filterAndRemovePrefixes(tags, rif.TagPrefix)
				}

				fmt.Printf("Without prefix\n%+v", tags)

				// if semver then sort first
				// (if not semver then will just get the last tag)
				latestVersion := tags[len(tags)-1]

				currentDependencies[remote.Url] = map[string]string{
					"constraint": currentVersion,
					"source":     "git",
					"repo":       remote.Url,
				}

				if latestVersion != currentVersion {
					updatedDependencies[remote.Url] = map[string]string{
						"constraint": latestVersion,
						"source":     "git",
						"repo":       remote.Url,
					}
				}

				// replace replacement
				// use replacement in
				// result := regex.ReplaceAllString(fileStr, rif.Pattern)
				// println(result)
			}
		}
	}

	// how do I want it to end up? has to be manifest to treat each individually by default
	// what if manifest name is ""? test it and see what happens
	// Update sentry in docs.md from / to /
	// Update sentry from / to /

	output := map[string]interface{}{
		"manifests": map[string]interface{}{
			"": map[string]interface{}{
				"current": map[string]interface{}{
					"dependencies": currentDependencies,
				},
				"updated": map[string]interface{}{
					"dependencies": updatedDependencies,
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

	// Preserve the order, but keep remove duplicates as we go
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
