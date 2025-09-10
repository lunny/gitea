// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pull

import (
	"context"

	git_model "code.gitea.io/gitea/models/git"
	issues_model "code.gitea.io/gitea/models/issues"
	repo_model "code.gitea.io/gitea/models/repo"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/json"
	git_service "code.gitea.io/gitea/services/git"
)

// getCommitIDsFromRepo get commit IDs from repo in between oldCommitID and newCommitID
// Commit on baseBranch will skip
func getCommitIDsFromRepo(ctx context.Context, repo *repo_model.Repository, gitRepo *git.Repository, oldCommitID, newCommitID, baseBranch string) ([]*git_model.SignCommitWithStatuses, error) {
	oldCommit, err := gitRepo.GetCommit(oldCommitID)
	if err != nil {
		return nil, err
	}

	newCommit, err := gitRepo.GetCommit(newCommitID)
	if err != nil {
		return nil, err
	}

	// Find commits between new and old commit excluding base branch commits
	commits, err := gitRepo.CommitsBetweenNotBase(newCommit, oldCommit, baseBranch)
	if err != nil {
		return nil, err
	}

	return git_service.ConvertFromGitCommit(ctx, commits, repo)
}

// CreatePushPullComment create push code to pull base comment
func CreatePushPullComment(ctx context.Context, gitRepo *git.Repository, pusher *user_model.User, pr *issues_model.PullRequest, oldCommitID, newCommitID string, isForcePush bool) (comment *issues_model.Comment, err error) {
	if pr.HasMerged || oldCommitID == "" || newCommitID == "" {
		return nil, nil
	}

	opts := &issues_model.CreateCommentOptions{
		Type:  issues_model.CommentTypePullRequestPush,
		Doer:  pusher,
		Repo:  pr.BaseRepo,
		Issue: pr.Issue,
	}

	data := issues_model.PushActionContent{
		IsForcePush: isForcePush,
	}
	if isForcePush {
		data.CommitIDs = []string{oldCommitID, newCommitID}
	} else {
		data.Commits, err = getCommitIDsFromRepo(ctx, pr.BaseRepo, gitRepo, oldCommitID, newCommitID, pr.BaseBranch)
		if err != nil {
			return nil, err
		}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	opts.Content = string(dataJSON)
	comment, err = issues_model.CreateComment(ctx, opts)

	return comment, err
}
