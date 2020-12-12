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
	BotTaskCanceled
	BotTaskTimeout
)

// BotTask represnets bot tasks
type BotTask struct {
	ID           int64
	RepoID       int64 `xorm:"index"`
	Ref string
	CommitID string
	Event        string
	EventPayload string `xorm:"LONGTEXT"`
	WorkerID     int64  `xorm:"index"`
	Status       BotTaskStatus
	Content      string             `xorm:"LONGTEXT"`
	Created      timeutil.TimeStamp `xorm:"created"`
	StartTime timeutil.TimeStamp
	EndTime timeutil.TimeStamp
	Updated      timeutil.TimeStamp `xorm:"updated"`
}

// InsertBotTask inserts a bot task
func InsertBotTask(t *BotTask) error {
	_, err := x.Insert(t)
	return err
}
