// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package code

import (
	"bytes"
	"html"
	gotemplate "html/template"
	"strings"

	"code.gitea.io/gitea/modules/highlight"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/modules/util"
)

// Result a search result to display
type Result struct {
	RepoID         int64
	Filename       string
	CommitID       string
	UpdatedUnix    timeutil.TimeStamp
	Language       string
	Color          string
	HighlightClass string
	LineNumbers    []int
	FormattedLines gotemplate.HTML
}

func indices(content string, selectionStartIndex, selectionEndIndex int) (int, int) {
	startIndex := selectionStartIndex
	numLinesBefore := 0
	for ; startIndex > 0; startIndex-- {
		if content[startIndex-1] == '\n' {
			if numLinesBefore == 1 {
				break
			}
			numLinesBefore++
		}
	}

	endIndex := selectionEndIndex
	numLinesAfter := 0
	for ; endIndex < len(content); endIndex++ {
		if content[endIndex] == '\n' {
			if numLinesAfter == 1 {
				break
			}
			numLinesAfter++
		}
	}

	return startIndex, endIndex
}

func writeStrings(buf *bytes.Buffer, strs ...string) error {
	for _, s := range strs {
		_, err := buf.WriteString(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchResult(result *SearchResult, startIndex, endIndex int) (*Result, error) {
	startLineNum := 1 + strings.Count(result.Content[:startIndex], "\n")

	var formattedLinesBuffer bytes.Buffer

	contentLines := strings.SplitAfter(result.Content[startIndex:endIndex], "\n")
	lineNumbers := make([]int, len(contentLines))
	index := startIndex
	positions := result.Positions

	for i, line := range contentLines {
		var err error

		err = writeStrings(&formattedLinesBuffer, `<li>`)
		if err != nil {
			return nil, err
		}

		pos := 0
		end := index + len(line)

		for len(positions) > 0 {
			p := &positions[0]
			if p.EndIndex <= p.StartIndex || p.EndIndex <= (index + pos) {
				positions = positions[1:]
				continue
			}

			if p.StartIndex >= end {
				break
			}

			openActiveIndex := util.Max(p.StartIndex-index, pos)
			closeActiveIndex := util.Min(p.EndIndex-index, len(line))

			err = writeStrings(&formattedLinesBuffer,
				html.EscapeString(line[pos:openActiveIndex]),
				`<span class='active'>`,
				html.EscapeString(line[openActiveIndex:closeActiveIndex]),
				`</span>`,
			)

			if err != nil {
				return nil, err
			}

			pos = closeActiveIndex
		}

		err = writeStrings(&formattedLinesBuffer,
			html.EscapeString(line[pos:]),
			`</li>`,
		)

		lineNumbers[i] = startLineNum + i
		index += len(line)
	}
	return &Result{
		RepoID:         result.RepoID,
		Filename:       result.Filename,
		CommitID:       result.CommitID,
		UpdatedUnix:    result.UpdatedUnix,
		Language:       result.Language,
		Color:          result.Color,
		HighlightClass: highlight.FileNameToHighlightClass(result.Filename),
		LineNumbers:    lineNumbers,
		FormattedLines: gotemplate.HTML(formattedLinesBuffer.String()),
	}, nil
}

// PerformSearch perform a search on a repository
func PerformSearch(repoIDs []int64, language, keyword string, page, pageSize int) (int, []*Result, []*SearchResultLanguages, error) {
	if len(keyword) == 0 {
		return 0, nil, nil, nil
	}

	total, results, resultLanguages, err := indexer.Search(repoIDs, language, keyword, page, pageSize)
	if err != nil {
		return 0, nil, nil, err
	}

	displayResults := make([]*Result, len(results))

	for i, result := range results {
		startIndex, endIndex := indices(result.Content, result.Positions[0].StartIndex, result.Positions[len(result.Positions)-1].EndIndex)
		displayResults[i], err = searchResult(result, startIndex, endIndex)
		if err != nil {
			return 0, nil, nil, err
		}
	}
	return int(total), displayResults, resultLanguages, nil
}
