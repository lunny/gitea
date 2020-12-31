// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bot

import (
	"encoding/json"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/bot/runner"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/notification/base"
	"code.gitea.io/gitea/modules/repository"
)

type botNotifier struct {
	base.NullNotifier
}

var (
	_ base.Notifier = &botNotifier{}
)

// NewNotifier create a new botNotifier notifier
func NewNotifier() base.Notifier {
	return &botNotifier{}
}

func notifyIssue(issue *models.Issue, evt models.HookEventType, payload string) {
	err := issue.LoadRepo()
	if err != nil {
		log.Error("issue.LoadRepo: %v", err)
		return
	}
	var ref = issue.Ref
	if ref == "" {
		ref = issue.Repo.DefaultBranch
	}

	gitRepo, err := git.OpenRepository(issue.Repo.RepoPath())
	if err != nil {
		log.Error("issue.LoadRepo: %v", err)
		return
	}
	defer gitRepo.Close()

	// Get the commit object for the ref
	commit, err := gitRepo.GetCommit(ref)
	if err != nil {
		log.Error("issue.LoadRepo: %v", err)
		return
	}

	runnerTypes := runner.GetRunnerTypes()
	for _, rt := range runnerTypes {
		found, content, err := rt.Detect(commit, evt, ref)
		if err != nil {
			log.Error("wt.Detect: %v", err)
			return
		}
		if !found {
			continue
		}

		if err := models.InsertBotTask(&models.BotTask{
			RepoID:       issue.RepoID,
			Type:         rt.Name(),
			Event:        string(evt),
			EventPayload: payload,
			Status:       models.BotTaskPending,
			Content:      content,
			Ref:          ref,
			CommitID:     commit.ID.String(),
		}); err != nil {
			log.Error("InsertBotTask: %v", err)
		}
	}
}

// TODO: implement all events
func (a *botNotifier) NotifyNewIssue(issue *models.Issue) {
	var payload = map[string]interface{}{
		"issue_index": issue.Index,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		log.Error("NotifyNewIssue: %v", err)
		return
	}
	notifyIssue(issue, models.HookEventIssues, string(bs))
}

// NotifyIssueChangeStatus notifies close or reopen issue to notifiers
func (a *botNotifier) NotifyIssueChangeStatus(doer *models.User, issue *models.Issue, actionComment *models.Comment, closeOrReopen bool) {
}

func (a *botNotifier) NotifyIssueChangeLabels(doer *models.User, issue *models.Issue,
	addedLabels []*models.Label, removedLabels []*models.Label) {
	var payload = map[string]interface{}{
		"issue_index": issue.Index,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		log.Error("NotifyNewIssue: %v", err)
		return
	}
	notifyIssue(issue, models.HookEventIssueLabel, string(bs))
}

// NotifyCreateIssueComment notifies comment on an issue to notifiers
func (a *botNotifier) NotifyCreateIssueComment(doer *models.User, repo *models.Repository,
	issue *models.Issue, comment *models.Comment, mentions []*models.User) {
}

func (a *botNotifier) NotifyNewPullRequest(pull *models.PullRequest, mentions []*models.User) {
}

func (a *botNotifier) NotifyRenameRepository(doer *models.User, repo *models.Repository, oldRepoName string) {
}

func (a *botNotifier) NotifyTransferRepository(doer *models.User, repo *models.Repository, oldOwnerName string) {

}

func (a *botNotifier) NotifyCreateRepository(doer *models.User, u *models.User, repo *models.Repository) {
}

func (a *botNotifier) NotifyForkRepository(doer *models.User, oldRepo, repo *models.Repository) {
}

func (a *botNotifier) NotifyPullRequestReview(pr *models.PullRequest, review *models.Review, comment *models.Comment, mentions []*models.User) {
}

func (*botNotifier) NotifyMergePullRequest(pr *models.PullRequest, doer *models.User) {
}

func (a *botNotifier) NotifyPushCommits(pusher *models.User, repo *models.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {

}

func (a *botNotifier) NotifyCreateRef(doer *models.User, repo *models.Repository, refType, refFullName string) {

}

func (a *botNotifier) NotifyDeleteRef(doer *models.User, repo *models.Repository, refType, refFullName string) {

}

func (a *botNotifier) NotifySyncPushCommits(pusher *models.User, repo *models.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {

}

func (a *botNotifier) NotifySyncCreateRef(doer *models.User, repo *models.Repository, refType, refFullName string) {

}

func (a *botNotifier) NotifySyncDeleteRef(doer *models.User, repo *models.Repository, refType, refFullName string) {

}

func (a *botNotifier) NotifyNewRelease(rel *models.Release) {

}
