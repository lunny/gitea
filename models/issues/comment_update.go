// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package issues

import (
	"context"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/json"
)

// UpdatePushCommentData updates the push comment data in the database.
func UpdatePushCommentData(ctx context.Context, c *Comment) error {
	data, err := json.Marshal(c.PushActionContent)
	if err != nil {
		return err
	}
	c.Content = string(data)
	_, err = db.GetEngine(ctx).ID(c.ID).Cols("content").NoAutoTime().Update(c)
	return err
}
