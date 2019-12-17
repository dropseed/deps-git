package main

import (
	"regexp"
)

type tagFilter struct {
	Matching string `json:"matching"`
	OutputAs string `json:"output_as"`
	SortAs   string `json:"sort_as"`
}

func (tf *tagFilter) processTag(t *Tag) {
	regex := regexp.MustCompile(tf.Matching)

	outputTemplate := tf.OutputAs
	if outputTemplate == "" {
		outputTemplate = "$0" // default
	}
	sortTemplate := tf.SortAs
	if sortTemplate == "" {
		sortTemplate = outputTemplate // default to same as output template
	}

	outputResult := []byte{}
	sortResult := []byte{}

	for _, submatches := range regex.FindAllStringSubmatchIndex(t.original, -1) {
		// Apply the captured submatches to the template and append the output
		// to the result.
		outputResult = regex.ExpandString(outputResult, outputTemplate, t.original, submatches)
		sortResult = regex.ExpandString(sortResult, sortTemplate, t.original, submatches)
	}

	t.outputAs = string(outputResult)
	t.sortAs = string(sortResult)
}
