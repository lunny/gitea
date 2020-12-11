// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import "code.gitea.io/gitea/modules/timeutil"

func init() {
	tables = append(tables, new(BotWorker), new(BotTask))
}

// BotWorker represents worker machines
type BotWorker struct {
	ID          int64
	Name        string
	Description string `xorm:"TEXT"`
	Token       string
	LastOnline  timeutil.TimeStamp
	Created     timeutil.TimeStamp `xorm:"created"`
}

// BotTaskStatus represents a task status
type BotTaskStatus int

// enumerate all the statuses of bot task
const (
	BotTaskPending BotTaskStatus = iota
	BotTaskRunning
	BotTaskFailed
	BotTaskFinished
)

// BotTask represnets bot tasks
type BotTask struct {
	ID       int64
	Event    string
	WorkerID int64 `xorm:"index"`
	RepoID   int64 `xorm:"index"`
	Branch   string
	PullID   int64
	Status   BotTaskStatus
	Content  string             `xorm:"BIGTEXT"`
	Created  timeutil.TimeStamp `xorm:"created"`
	Updated  timeutil.TimeStamp `xorm:"updated"`
}

// InsertBotTask inserts a bot task
func InsertBotTask(t *BotTask) error {
	_, err := x.Insert(t)
	return err
}
