// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"path"

	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/storage"

	"xorm.io/xorm"
)

func relativePath(uuid string) string {
	return path.Join(uuid[0:1], uuid[1:2], uuid)
}

func deleteOrphanedAttachments(x *xorm.Engine) error {
	type Attachment struct {
		ID        int64  `xorm:"pk autoincr"`
		UUID      string `xorm:"uuid UNIQUE"`
		IssueID   int64  `xorm:"INDEX"`
		ReleaseID int64  `xorm:"INDEX"`
		CommentID int64
	}

	sess := x.NewSession()
	defer sess.Close()

	var limit = setting.Database.IterateBufferSize
	if limit <= 0 {
		limit = 50
	}

	for {
		attachements := make([]Attachment, 0, limit)
		if err := sess.Where("`issue_id` = 0 and (`release_id` = 0 or `release_id` not in (select `id` from `release`))").
			Cols("id, uuid").Limit(limit).
			Asc("id").
			Find(&attachements); err != nil {
			return err
		}
		if len(attachements) == 0 {
			return nil
		}

		var ids = make([]int64, 0, limit)
		for _, attachment := range attachements {
			ids = append(ids, attachment.ID)
		}
		if len(ids) > 0 {
			if _, err := sess.In("id", ids).Delete(new(Attachment)); err != nil {
				return err
			}
		}

		for _, attachment := range attachements {
			if err := storage.Attachments.Delete(relativePath(attachment.UUID)); err != nil {
				return err
			}
		}
		if len(attachements) < limit {
			return nil
		}
	}
}
