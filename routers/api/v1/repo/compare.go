// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"errors"
	"net/http"

	access_model "code.gitea.io/gitea/models/perm/access"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unit"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/gitrepo"
	"code.gitea.io/gitea/modules/log"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/routers/common"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/convert"
)

// CompareDiff compare two branches or commits
func CompareDiff(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/compare/{basehead} repository repoCompareDiff
	// ---
	// summary: Get commit comparison information
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: basehead
	//   in: path
	//   description: compare two branches or commits
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Compare"
	//   "404":
	//     "$ref": "#/responses/notFound"

	if ctx.Repo.GitRepo == nil {
		gitRepo, err := gitrepo.OpenRepository(ctx, ctx.Repo.Repository)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "OpenRepository", err)
			return
		}
		ctx.Repo.GitRepo = gitRepo
		defer gitRepo.Close()
	}

	pathParam := ctx.PathParam("*")
	baseRepo := ctx.Repo.Repository
	ci, err := common.ParseComparePathParams(ctx, pathParam, baseRepo, ctx.Repo.GitRepo)
	if err != nil {
		switch {
		case user_model.IsErrUserNotExist(err):
			ctx.NotFound("GetUserByName")
		case repo_model.IsErrRepoNotExist(err):
			ctx.NotFound("GetRepositoryByOwnerAndName")
		case errors.Is(err, util.ErrInvalidArgument):
			ctx.NotFound("ParseComparePathParams")
		default:
			ctx.ServerError("GetRepositoryByOwnerAndName", err)
		}
		return
	}
	defer ci.Close()

	// remove the check when we support compare with carets
	if ci.CaretTimes > 0 {
		ctx.NotFound("Unsupported compare")
		return
	}

	if !ci.IsSameRepo() {
		// user should have permission to read headrepo's codes
		permHead, err := access_model.GetUserRepoPermission(ctx, ci.HeadRepo, ctx.Doer)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "GetUserRepoPermission", err)
			return
		}
		if !permHead.CanRead(unit.TypeCode) {
			if log.IsTrace() {
				log.Trace("Permission Denied: User: %-v cannot read code in Repo: %-v\nUser in headRepo has Permissions: %-+v",
					ctx.Doer,
					ci.HeadRepo,
					permHead)
			}
			ctx.NotFound("Can't read headRepo UnitTypeCode")
			return
		}
	}

	ctx.Repo.PullRequest.SameRepo = ci.IsSameRepo()
	log.Trace("Repo path: %q, base branch: %q, head branch: %q", ctx.Repo.GitRepo.Path, ci.BaseOriRef, ci.HeadOriRef)

	ci.CompareInfo, err = ci.HeadGitRepo.GetCompareInfo(repo_model.RepoPath(baseRepo.Owner.Name, baseRepo.Name), ci.BaseOriRef, ci.HeadOriRef, false, false)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetCompareInfo", err)
		return
	}

	verification := ctx.FormString("verification") == "" || ctx.FormBool("verification")
	files := ctx.FormString("files") == "" || ctx.FormBool("files")

	apiCommits := make([]*api.Commit, 0, len(ci.CompareInfo.Commits))
	userCache := make(map[string]*user_model.User)
	for i := 0; i < len(ci.CompareInfo.Commits); i++ {
		apiCommit, err := convert.ToCommit(ctx, ctx.Repo.Repository, ctx.Repo.GitRepo, ci.CompareInfo.Commits[i], userCache,
			convert.ToCommitOptions{
				Stat:         true,
				Verification: verification,
				Files:        files,
			})
		if err != nil {
			ctx.ServerError("toCommit", err)
			return
		}
		apiCommits = append(apiCommits, apiCommit)
	}

	ctx.JSON(http.StatusOK, &api.Compare{
		TotalCommits: len(ci.CompareInfo.Commits),
		Commits:      apiCommits,
	})
}
