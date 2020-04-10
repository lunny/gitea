// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package code

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/analyze"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/charset"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/timeutil"

	"github.com/olivere/elastic/v7"
	"github.com/src-d/enry/v2"
)

var (
	_ Indexer = &ElasticSearchIndexer{}
)

// ElasticSearchIndexer implements Indexer interface
type ElasticSearchIndexer struct {
	client      *elastic.Client
	indexerName string
}

type elasticLogger struct {
	*log.Logger
}

func (l elasticLogger) Printf(format string, args ...interface{}) {
	_ = l.Logger.Log(2, l.Logger.GetLevel(), format, args...)
}

// NewElasticSearchIndexer creates a new elasticsearch indexer
func NewElasticSearchIndexer(url, indexerName string) (*ElasticSearchIndexer, bool, error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(url),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10 * time.Second),
		elastic.SetGzip(false),
	}

	logger := elasticLogger{log.GetLogger(log.DEFAULT)}

	if logger.GetLevel() == log.TRACE || logger.GetLevel() == log.DEBUG {
		opts = append(opts, elastic.SetTraceLog(logger))
	} else if logger.GetLevel() == log.ERROR || logger.GetLevel() == log.CRITICAL || logger.GetLevel() == log.FATAL {
		opts = append(opts, elastic.SetErrorLog(logger))
	} else if logger.GetLevel() == log.INFO || logger.GetLevel() == log.WARN {
		opts = append(opts, elastic.SetInfoLog(logger))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, false, err
	}

	indexer := &ElasticSearchIndexer{
		client:      client,
		indexerName: indexerName,
	}
	exists, err := indexer.init()

	return indexer, exists, err
}

const (
	defaultMapping = `{
		"mappings": {
			"properties": {
				"repo_id": {
					"type": "long",
					"index": true
				},
				"content": {
					"type": "text",
					"index": true
				},
				"commit_id": {
					"type": "keyword",
					"index": true
				},
				"language": {
					"type": "keyword",
					"index": true
				},
				"updated_at": {
					"type": "long",
					"index": true
				}
			}
		}
	}`
)

// Init will initialize the indexer
func (b *ElasticSearchIndexer) init() (bool, error) {
	ctx := context.Background()
	exists, err := b.client.IndexExists(b.indexerName).Do(ctx)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	var mapping = defaultMapping

	createIndex, err := b.client.CreateIndex(b.indexerName).BodyString(mapping).Do(ctx)
	if err != nil {
		return false, err
	}
	if !createIndex.Acknowledged {
		return false, errors.New("init failed")
	}

	return false, nil
}

func (b *ElasticSearchIndexer) addUpdate(sha string, update fileUpdate, repo *models.Repository) ([]elastic.BulkableRequest, error) {
	stdout, err := git.NewCommand("cat-file", "-s", update.BlobSha).
		RunInDir(repo.RepoPath())
	if err != nil {
		return nil, err
	}
	if size, err := strconv.Atoi(strings.TrimSpace(stdout)); err != nil {
		return nil, fmt.Errorf("Misformatted git cat-file output: %v", err)
	} else if int64(size) > setting.Indexer.MaxIndexerFileSize {
		return b.addDelete(update.Filename, repo)
	}

	fileContents, err := git.NewCommand("cat-file", "blob", update.BlobSha).
		RunInDirBytes(repo.RepoPath())
	if err != nil {
		return nil, err
	} else if !base.IsTextFile(fileContents) {
		// FIXME: UTF-16 files will probably fail here
		return nil, nil
	}

	id := filenameIndexerID(repo.ID, update.Filename)

	return []elastic.BulkableRequest{
		elastic.NewBulkIndexRequest().
			Index(b.indexerName).
			Id(id).
			Doc(map[string]interface{}{
				"repo_id":    repo.ID,
				"content":    string(charset.ToUTF8DropErrors(fileContents)),
				"commit_id":  sha,
				"language":   analyze.GetCodeLanguage(update.Filename, fileContents),
				"updated_at": timeutil.TimeStampNow(),
			}),
	}, nil
}

func (b *ElasticSearchIndexer) addDelete(filename string, repo *models.Repository) ([]elastic.BulkableRequest, error) {
	id := filenameIndexerID(repo.ID, filename)
	return []elastic.BulkableRequest{
		elastic.NewBulkDeleteRequest().
			Index(b.indexerName).
			Id(id),
	}, nil
}

// Index will save the index data
func (b *ElasticSearchIndexer) Index(repo *models.Repository, sha string, changes *repoChanges) error {
	reqs := make([]elastic.BulkableRequest, 0)
	for _, update := range changes.Updates {
		updateReqs, err := b.addUpdate(sha, update, repo)
		if err != nil {
			return err
		}
		if len(updateReqs) > 0 {
			reqs = append(reqs, updateReqs...)
		}
	}

	for _, filename := range changes.RemovedFilenames {
		delReqs, err := b.addDelete(filename, repo)
		if err != nil {
			return err
		}
		if len(delReqs) > 0 {
			reqs = append(reqs, delReqs...)
		}
	}

	if len(reqs) > 0 {
		_, err := b.client.Bulk().
			Index(b.indexerName).
			Add(reqs...).
			Do(context.Background())
		return err
	}
	return nil
}

// Delete deletes indexes by ids
func (b *ElasticSearchIndexer) Delete(repoID int64) error {
	_, err := b.client.DeleteByQuery(b.indexerName).
		Query(elastic.NewTermsQuery("repo_id", repoID)).
		Do(context.Background())
	return err
}

func convertResult(searchResult *elastic.SearchResult, kw string, pageSize int) (int64, []*SearchResult, []*SearchResultLanguages, error) {
	hits := make([]*SearchResult, 0, pageSize)
	for _, hit := range searchResult.Hits.Hits {
		repoID, fileName := parseIndexerID(hit.Id)
		var res = make(map[string]interface{})
		if err := json.Unmarshal(hit.Source, &res); err != nil {
			return 0, nil, nil, err
		}

		language := res["language"].(string)
		commitId := res["commit_id"].(string)
		content := res["content"].(string)
		updateUnix := timeutil.TimeStamp(res["updated_at"].(float64))
		color := enry.GetColor(language)

		c, ok := hit.Highlight["content"]
		if ok && len(c) > 0 {
			for _, p := range c {
				s := strings.FieldsFunc(p, func(r rune) bool { return r == ':' })

				if len(s) != 3 {
					continue
				}

				s = strings.FieldsFunc(s[1], func(r rune) bool { return r == ',' })

				pm := map[int]int{}
				pl := make([]SearchResultPosition, 0)

				for _, r := range s {
					rr := strings.FieldsFunc(r, func(r rune) bool { return r == '-' })
					if len(rr) != 2 {
						continue
					}

					start, e1 := strconv.Atoi(rr[0])
					end, e2 := strconv.Atoi(rr[1])
					if e1 != nil || e2 != nil {
						continue
					}

					pm[start] = -1
					pm[end] = -1

					pl = append(pl, SearchResultPosition{StartIndex: start, EndIndex: end})
				}

				if len(pl) == 0 {
					continue
				}

				idx := 0
				for cp, _ := range content {
					_, ok := pm[idx]
					if ok {
						pm[idx] = cp
					}
					idx++
				}

				for i := range pl {
					p := &pl[i]
					p.StartIndex = pm[p.StartIndex]
					p.EndIndex = pm[p.EndIndex]
				}

				hits = append(hits, &SearchResult{
					RepoID:      repoID,
					Filename:    fileName,
					CommitID:    commitId,
					Content:     content,
					UpdatedUnix: updateUnix,
					Language:    language,
					Positions:   pl,
					Color:       color,
				})
			}
		}
	}

	return searchResult.TotalHits(), hits, extractAggs(searchResult), nil
}

func extractAggs(searchResult *elastic.SearchResult) []*SearchResultLanguages {
	var searchResultLanguages []*SearchResultLanguages
	agg, found := searchResult.Aggregations.Terms("language")
	if found {
		searchResultLanguages = make([]*SearchResultLanguages, 0, 10)

		for _, bucket := range agg.Buckets {
			searchResultLanguages = append(searchResultLanguages, &SearchResultLanguages{
				Language: bucket.Key.(string),
				Color:    enry.GetColor(bucket.Key.(string)),
				Count:    int(bucket.DocCount),
			})
		}
	}
	return searchResultLanguages
}

// Search searches for codes and language stats by given conditions.
func (b *ElasticSearchIndexer) Search(repoIDs []int64, language, keyword string, page, pageSize int) (int64, []*SearchResult, []*SearchResultLanguages, error) {
	kwQuery := elastic.NewQueryStringQuery(keyword).
		Field("content").
		Fuzziness("AUTO").
		AnalyzeWildcard(true).
		Lenient(true)

	query := elastic.NewBoolQuery()
	query = query.Must(kwQuery)
	if len(repoIDs) > 0 {
		var repoStrs = make([]interface{}, 0, len(repoIDs))
		for _, repoID := range repoIDs {
			repoStrs = append(repoStrs, repoID)
		}
		repoQuery := elastic.NewTermsQuery("repo_id", repoStrs...)
		query = query.Must(repoQuery)
	}

	var (
		start       int
		kw          = "<em>" + keyword + "</em>"
		aggregation = elastic.NewTermsAggregation().Field("language").Size(10).OrderByCountDesc()
	)

	if page > 0 {
		start = (page - 1) * pageSize
	}

	highlight := elastic.NewHighlight().
		Fields(elastic.NewHighlighterField("content").
			HighlighterType("experimental").
			Options(map[string]interface{}{"return_offsets": true}))

	if len(language) == 0 {
		searchResult, err := b.client.Search().
			Index(b.indexerName).
			Aggregation("language", aggregation).
			Query(query).
			Highlight(highlight).
			Sort("repo_id", true).
			From(start).Size(pageSize).
			Do(context.Background())
		if err != nil {
			return 0, nil, nil, err
		}

		return convertResult(searchResult, kw, pageSize)
	}

	langQuery := elastic.NewMatchQuery("language", language)
	countResult, err := b.client.Search().
		Index(b.indexerName).
		Aggregation("language", aggregation).
		Query(query).
		Size(0). // We only needs stats information
		Do(context.Background())
	if err != nil {
		return 0, nil, nil, err
	}

	query = query.Must(langQuery)
	searchResult, err := b.client.Search().
		Index(b.indexerName).
		Query(query).
		Highlight(highlight).
		Sort("repo_id", true).
		From(start).Size(pageSize).
		Do(context.Background())
	if err != nil {
		return 0, nil, nil, err
	}

	total, hits, _, err := convertResult(searchResult, kw, pageSize)

	return total, hits, extractAggs(countResult), err
}

// Close implements indexer
func (b *ElasticSearchIndexer) Close() {}
