// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"strings"
	"time"

	"code.gitea.io/gitea/modules/log"
)

// Cache represents cache settings
type Cache struct {
	Enabled  bool
	Adapter  string
	Interval int
	Conn     string
	TTL      time.Duration `ini:"ITEM_TTL"`
}

var (
	// CacheService the global cache
	CacheService = struct {
		Cache

		LastCommit struct {
			Cache
			CommitsCount int64
		} `ini:"cache.last_commit"`
	}{
		Cache: Cache{
			Enabled:  true,
			Adapter:  "memory",
			Interval: 60,
			TTL:      16 * time.Hour,
		},
		LastCommit: struct {
			Cache
			CommitsCount int64
		}{
			Cache: Cache{
				Enabled:  true,
				Adapter:  "memory",
				Interval: 60,
				TTL:      86400 * time.Hour,
			},
			CommitsCount: 1000,
		},
	}
)

func newCacheService() {
	sec := Cfg.Section("cache")
	if err := sec.MapTo(&CacheService); err != nil {
		log.Fatal("Failed to map Cache settings: %v", err)
	}

	CacheService.Adapter = sec.Key("ADAPTER").In("memory", []string{"memory", "redis", "memcache"})
	switch CacheService.Adapter {
	case "memory":
	case "redis", "memcache":
		CacheService.Conn = strings.Trim(sec.Key("HOST").String(), "\" ")
	case "": // disable cache
		CacheService.Enabled = false
	default:
		log.Fatal("Unknown cache adapter: %s", CacheService.Adapter)
	}

	if CacheService.Enabled {
		log.Info("Cache Service Enabled")
	}

	sec = Cfg.Section("cache.last_commit")

	CacheService.LastCommit.Adapter = sec.Key("ADAPTER").In("memory", []string{"memory", "redis", "memcache"})
	switch CacheService.LastCommit.Adapter {
	case "memory":
	case "redis", "memcache":
		CacheService.LastCommit.Conn = strings.Trim(sec.Key("HOST").String(), "\" ")
	case "": // disable cache
		CacheService.LastCommit.Enabled = false
	default:
		log.Fatal("Unknown cache.last_commit adapter: %s", CacheService.LastCommit.Adapter)
	}

	CacheService.LastCommit.CommitsCount = sec.Key("COMMITS_COUNT").MustInt64(1000)

	if CacheService.LastCommit.Enabled {
		log.Info("Last Commit Cache Service Enabled")
	}
}
