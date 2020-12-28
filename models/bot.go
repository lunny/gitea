// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"

	"code.gitea.io/gitea/modules/timeutil"
)

func init() {
	tables = append(tables, new(BotWorker), new(BotTask))
}

// ErrBotWorkerNotExist represents an error for bot worker not exist
type ErrBotWorkerNotExist struct {
	UUID string
}

func (err ErrBotWorkerNotExist) Error() string {
	return fmt.Sprintf("Bot worker [%s] is not exist", err.UUID)
}

// BotWorker represents worker machines
type BotWorker struct {
	ID          int64
	UUID        string `xorm:"CHAR(36) UNIQUE"`
	Name        string `xorm:"VARCHAR(32) UNIQUE"`
	OrgID       int64  `xorm:"index"` // org level worker
	RepoID      int64  `xorm:"index"` // repo level worker, if orgid also is zero, then it's a global
	Description string `xorm:"TEXT"`
	Token       string
	LastOnline  timeutil.TimeStamp
	Created     timeutil.TimeStamp `xorm:"created"`
}

// GetBotWorkerByUUID returns a bot worker via uuid
func GetBotWorkerByUUID(uuid string) (*BotWorker, error) {
	var worker BotWorker
	has, err := x.Where("uuid=?", uuid).Get(&worker)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrBotWorkerNotExist{
			UUID: uuid,
		}
	}
	return &worker, nil
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
	Ref          string
	CommitID     string
	Event        string
	Token        string // token for this task
	Grant        string // permissions for this task
	EventPayload string `xorm:"LONGTEXT"`
	WorkerID     int64  `xorm:"index"`
	Status       BotTaskStatus
	Content      string             `xorm:"LONGTEXT"`
	Created      timeutil.TimeStamp `xorm:"created"`
	StartTime    timeutil.TimeStamp
	EndTime      timeutil.TimeStamp
	Updated      timeutil.TimeStamp `xorm:"updated"`
}

// InsertBotTask inserts a bot task
func InsertBotTask(t *BotTask) error {
	_, err := x.Insert(t)
	return err
}

// GetBotTasks returns all the tasks for the bot
func GetBotTasks(botID int64) ([]BotTask, error) {
	var tasks []BotTask
	// FIXME: for test, just return all tasks
	err := x.Where("status=?", BotTaskPending).Find(&tasks)
	//err := x.Where("worker_id = ?", botID).
	// And("status=?", BotTaskPending).
	// Find(&tasks)
	return tasks, err
}
