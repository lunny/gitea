// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"

	"code.gitea.io/gitea/modules/timeutil"
	uuid "github.com/google/uuid"
)

func init() {
	tables = append(tables, new(BotRunner), new(BotTask))
}

// ErrBotRunnerNotExist represents an error for bot runner not exist
type ErrBotRunnerNotExist struct {
	UUID string
}

func (err ErrBotRunnerNotExist) Error() string {
	return fmt.Sprintf("Bot runner [%s] is not exist", err.UUID)
}

// BotRunner represents runner machines
type BotRunner struct {
	ID          int64
	UUID        string `xorm:"CHAR(36) UNIQUE"`
	Name        string `xorm:"VARCHAR(32) UNIQUE"`
	Type        string `xorm:"VARCHAR(16)"`
	OwnerID     int64  `xorm:"index"` // org level runner
	RepoID      int64  `xorm:"index"` // repo level runner, if orgid also is zero, then it's a global
	Description string `xorm:"TEXT"`
	Token       string
	LastOnline  timeutil.TimeStamp
	Created     timeutil.TimeStamp `xorm:"created"`
}

// GetBotRunnerByUUID returns a bot runner via uuid
func GetBotRunnerByUUID(uuid string) (*BotRunner, error) {
	var runner BotRunner
	has, err := x.Where("uuid=?", uuid).Get(&runner)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrBotRunnerNotExist{
			UUID: uuid,
		}
	}
	return &runner, nil
}

// FindRunnersByRepoID returns all workers for the repository
func FindRunnersByRepoID(repoID int64) ([]*BotRunner, error) {
	var runners []*BotRunner
	err := x.Where("repo_id=? OR repo_id=0", repoID).
		Find(&runners)
	if err != nil {
		return nil, err
	}
	err = x.Join("INNER", "repository", "repository.owner_id = bot_worker.owner_id").Find(&runners)
	return runners, err
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
	UUID         string `xorm:"CHAR(36)"`
	RepoID       int64  `xorm:"index"`
	Type         string `xorm:"VARCHAR(16)"`
	Ref          string
	CommitID     string
	Event        string
	Token        string // token for this task
	Grant        string // permissions for this task
	EventPayload string `xorm:"LONGTEXT"`
	RunnerID     int64  `xorm:"index"`
	Status       BotTaskStatus
	Content      string             `xorm:"LONGTEXT"`
	Created      timeutil.TimeStamp `xorm:"created"`
	StartTime    timeutil.TimeStamp
	EndTime      timeutil.TimeStamp
	Updated      timeutil.TimeStamp `xorm:"updated"`
}

// InsertBotTask inserts a bot task
func InsertBotTask(t *BotTask) error {
	if t.UUID == "" {
		t.UUID = uuid.New().String()
	}
	_, err := x.Insert(t)
	return err
}

// UpdateBotTask updates bot task
func UpdateBotTask(t *BotTask, cols ...string) error {
	_, err := x.ID(t.ID).Cols(cols...).Update(t)
	return err
}

// ErrBotTaskNotExist represents an error for bot task not exist
type ErrBotTaskNotExist struct {
	UUID string
}

func (err ErrBotTaskNotExist) Error() string {
	return fmt.Sprintf("Bot task [%s] is not exist", err.UUID)
}

// GetBotTaskByUUID gets bot task by uuid
func GetBotTaskByUUID(botTaskUUID string) (*BotTask, error) {
	var task BotTask
	has, err := x.Where("uuid=?", botTaskUUID).Get(&task)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrBotTaskNotExist{
			UUID: botTaskUUID,
		}
	}
	return &task, nil
}

// GetBotTask return the task for the bot
func GetBotTask(botRunnerID int64) (*BotTask, error) {
	var tasks []BotTask
	// FIXME: for test, just return all tasks
	err := x.Where("status=?", BotTaskPending).Find(&tasks)
	//err := x.Where("runner_id = ?", botID).
	// And("status=?", BotTaskPending).
	// Find(&tasks)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return &tasks[0], err
}
