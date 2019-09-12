package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
)

type remote struct {
	ReplaceInFiles []*replaceInFile `json:"replace_in_files"`
}

type replaceInFile struct {
	Filename    string `json:"filename"`
	Pattern     string `json:"pattern"`
	TagPrefix   string `json:"tag_prefix"`
	Semver      *bool  `json:"semver"`
	Prereleases bool   `json:"prereleases"`
	// could allow user to specify a range
}

func loadRemotesFromEnv() map[string]*remote {
	var remotes map[string]*remote
	if remotesStr := os.Getenv("DEPS_SETTING_REMOTES"); remotesStr != "" {
		if err := json.Unmarshal([]byte(remotesStr), &remotes); err != nil {
			panic(err)
		}
	}
	return remotes
}

func (rif *replaceInFile) regex() *regexp.Regexp {
	return regexp.MustCompile(rif.Pattern)
}

func (rif *replaceInFile) readFile() string {
	fileBytes, err := ioutil.ReadFile(rif.Filename)
	if err != nil {
		panic(err)
	}
	return string(fileBytes)
}

func (rif *replaceInFile) writeFile(contents string) {
	info, err := os.Stat(rif.Filename)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(rif.Filename, []byte(contents), info.Mode()); err != nil {
		panic(err)
	}
}

func (rif *replaceInFile) getLatestTag(tags []string) string {
	if rif.TagPrefix != "" {
		fmt.Printf("Filtering to tags with prefix %s and removing it", rif.TagPrefix)
		tags = filterAndRemovePrefixes(tags, rif.TagPrefix)
		fmt.Printf("Remaining tags: %v\n", tags)
	}

	// Enabled if not set
	if rif.Semver == nil || *rif.Semver {
		versions := stringsToVersions(tags, rif.Prereleases)
		if len(versions) < 1 && rif.TagPrefix == "" {
			// Try automatically removing "v" since it's so common
			tags = filterAndRemovePrefixes(tags, "v")
			versions = stringsToVersions(tags, rif.Prereleases)
		}
		tags = versionsToStrings(versions)
	}

	if len(tags) < 1 {
		return ""
	}

	latestVersion := tags[len(tags)-1]
	return latestVersion
}

func stringsToVersions(strs []string, includePrereleases bool) semver.Versions {
	versions := semver.Versions{}
	for _, s := range strs {
		version, err := semver.Make(s)
		if err != nil {
			continue
		}
		if len(version.Pre) > 0 && !includePrereleases {
			continue
		}
		versions = append(versions, version)
	}
	sort.Sort(versions)
	return versions
}

func versionsToStrings(versions semver.Versions) []string {
	strs := []string{}
	for _, v := range versions {
		strs = append(strs, v.String())
	}
	return strs
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
