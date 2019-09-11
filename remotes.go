package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
)

type remote struct {
	ReplaceInFiles []*replaceInFile `json:"replace_in_files"`
}

type replaceInFile struct {
	Filename  string `json:"filename"`
	Pattern   string `json:"pattern"`
	TagPrefix string `json:"tag_prefix"`
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
