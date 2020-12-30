// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bot

import (
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/convert"
)

// GetTask returns tasks to a bot
func GetTask(ctx *context.APIContext) {
	/*uuid := ctx.Query("bot_uuid")
	bot, err := models.GetBotRunnerByUUID(uuid)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}*/
	task, err := models.GetBotTask(0 /*bot.ID*/)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	if task == nil {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"task": nil,
		})
		return
	}

	repo, err := models.GetRepositoryByID(task.RepoID)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"task": convert.ToBotTask(repo, task),
	})
}

// UpdateTask updates task status
func UpdateTask(ctx *context.APIContext) {
	uuid := ctx.Params("uuid")
	status := ctx.Query("status")
	task, err := models.GetBotTaskByUUID(uuid)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	repo, err := models.GetRepositoryByID(task.RepoID)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	if status == "1" {
		task.Status = models.BotTaskFinished
	} else {
		task.Status = models.BotTaskFailed
	}
	if err := models.UpdateBotTask(task); err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"task": convert.ToBotTask(repo, task),
	})
}

// UploadLog allows bot runner to upload logs
func UploadLog(ctx *context.APIContext) {

}
