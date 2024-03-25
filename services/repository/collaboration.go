// Copyright 2016 The Gogs Authors. All rights reserved.
// Copyright 2020 The Gitea Authors.
// SPDX-License-Identifier: MIT

package repository

import (
	"context"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	access_model "code.gitea.io/gitea/models/perm/access"
	repo_model "code.gitea.io/gitea/models/repo"
	user_model "code.gitea.io/gitea/models/user"
)

// DeleteCollaboration removes collaboration relation between the user and repository.
func DeleteCollaboration(ctx context.Context, repo *repo_model.Repository, collaborator *user_model.User) (err error) {
	collaboration := &repo_model.Collaboration{
		RepoID: repo.ID,
		UserID: collaborator.ID,
	}

	return db.WithTx(ctx, func(ctx context.Context) error {
		if has, err := db.GetEngine(ctx).Delete(collaboration); err != nil {
			return err
		} else if has == 0 {
			return nil
		}

		if err := repo.LoadOwner(ctx); err != nil {
			return err
		}

		if err = access_model.RecalculateAccesses(ctx, repo); err != nil {
			return err
		}

		if err = repo_model.WatchRepo(ctx, collaborator, repo, false); err != nil {
			return err
		}

		if err = models.ReconsiderWatches(ctx, repo, collaborator); err != nil {
			return err
		}

		// Unassign a user from any issue (s)he has been assigned to in the repository
		return models.ReconsiderRepoIssuesAssignee(ctx, repo, collaborator)
	})
}
