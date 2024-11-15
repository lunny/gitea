// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git_test

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	git_model "code.gitea.io/gitea/models/git"
	"code.gitea.io/gitea/models/unittest"

	"github.com/stretchr/testify/assert"
)

func TestFindBranchesByRepoAndBranchName(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	// With no repos or branches given, we find no branches.
	branches, err := git_model.FindBranchesByRepoAndBranchName(db.DefaultContext, map[int64]string{})
	assert.NoError(t, err)
	assert.Len(t, branches, 0)

	// With a repo but branches given.
	branches, err = git_model.FindBranchesByRepoAndBranchName(db.DefaultContext, map[int64]string{
		1:  "foo",
		11: "master",
	})
	assert.NoError(t, err)
	assert.Len(t, branches, 2)
}
