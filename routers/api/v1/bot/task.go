// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bot

import (
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
)

// GetTask returns tasks to a bot
func GetTask(ctx *context.APIContext) {
	/*uuid := ctx.Query("bot_uuid")
	bot, err := models.GetBotWorkerByUUID(uuid)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}*/
	tasks, err := models.GetBotTasks(0 /*bot.ID*/)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"tasks": tasks,
	})
}

// UpdateTask updates task status
func UpdateTask(ctx *context.APIContext) {

}

// UploadLog allows bot worker to upload logs
func UploadLog(ctx *context.APIContext) {

}
