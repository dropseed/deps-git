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
	Filename    string     `json:"filename"`
	Pattern     string     `json:"pattern"`
	TagPrefix   string     `json:"tag_prefix"` // deprecated
	TagFilter   *tagFilter `json:"tag_filter"`
	Semver      *bool      `json:"semver"`
	Prereleases bool       `json:"prereleases"`
	Range       string     `json:"range"`
}

type Tag struct {
	original string
	outputAs string
	sortAs   string
	semver   *semver.Version
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

func (rif *replaceInFile) getLatestTag(tagStrs []string) string {

	if rif.TagPrefix != "" && rif.TagFilter != nil {
		panic("Cannot use tag_prefix and tag_filter together. Please choose one.")
	}

	if rif.TagPrefix != "" {
		fmt.Println("DEPRECATED: tag_prefix will be removed in version 1.0")
	}

	tags := []*Tag{}

	for _, t := range tagStrs {
		tag := &Tag{
			original: t,
			outputAs: t,
			sortAs:   t,
		}

		if rif.TagPrefix != "" {
			rif.TagFilter = &tagFilter{
				Matching: "^" + rif.TagPrefix + "(.*)",
				OutputAs: "$1",
				SortAs:   "$1",
			}
		}

		if rif.TagFilter != nil {
			rif.TagFilter.processTag(tag)
			fmt.Printf("Tag p: %s    output as: %s     sort as: %s\n", tag.original, tag.outputAs, tag.sortAs)
			if tag.outputAs == "" || tag.sortAs == "" {
				// Disregard this tag if we didn't get replacements
				continue
			}
		}

		if rif.Semver == nil || *rif.Semver {
			tag.semver = stringToSemver(tag.sortAs, rif.Range, rif.Prereleases)

			// A "v" prefix is so common that we'll try it automatically
			if tag.semver == nil && strings.HasPrefix(tag.sortAs, "v") {
				tag.semver = stringToSemver(tag.sortAs[1:], rif.Range, rif.Prereleases)
				if tag.semver != nil {
					// We'll remove the v from the output as well
					// as a part of this automatic feature
					tag.outputAs = tag.sortAs[1:]
				}
			}

			if tag.semver == nil {
				// This version was not compatible semver
				continue
			}
		}

		fmt.Printf("Tag original: %s    output as: %s     sort as: %s\n", tag.original, tag.outputAs, tag.sortAs)
		tags = append(tags, tag)
	}

	// only sorting we do right now is semver-based
	if rif.Semver == nil || *rif.Semver {
		sort.SliceStable(tags, func(i, j int) bool {
			return tags[i].semver.LT(*tags[j].semver)
		})
	}

	if len(tags) < 1 {
		return ""
	}

	latestVersion := tags[len(tags)-1]
	return latestVersion.outputAs
}

func stringToSemver(s string, rangeStr string, includePrereleases bool) *semver.Version {
	version, err := semver.Make(s)
	if err != nil {
		// Not a valid semver
		return nil
	}

	if len(version.Pre) > 0 && !includePrereleases {
		// This is a pre-release and they aren't included
		return nil
	}

	var semverRange semver.Range
	if rangeStr != "" {
		semverRange = semver.MustParseRange(rangeStr)
	}
	if semverRange != nil && !semverRange(version) {
		// There's a range and it's not in range
		return nil
	}

	return &version
}
