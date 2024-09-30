// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/log"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/tests"

	"github.com/stretchr/testify/assert"
)

// TestAPICreateAndDeleteToken tests that token that was just created can be deleted
func TestAPICreateAndDeleteToken(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	newAccessToken := createAPIAccessTokenWithoutCleanUp(t, "test-key-1", user, []auth_model.AccessTokenScope{auth_model.AccessTokenScopeAll})
	deleteAPIAccessToken(t, newAccessToken, user)

	newAccessToken = createAPIAccessTokenWithoutCleanUp(t, "test-key-2", user, []auth_model.AccessTokenScope{auth_model.AccessTokenScopeAll})
	deleteAPIAccessToken(t, newAccessToken, user)
}

// TestAPIDeleteMissingToken ensures that error is thrown when token not found
func TestAPIDeleteMissingToken(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	req := NewRequestf(t, "DELETE", "/api/v1/users/user1/tokens/%d", unittest.NonexistentID).
		AddBasicAuth(user.Name)
	MakeRequest(t, req, http.StatusNotFound)
}

// TestAPIGetTokensPermission ensures that only the admin can get tokens from other users
func TestAPIGetTokensPermission(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// admin can get tokens for other users
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	req := NewRequest(t, "GET", "/api/v1/users/user2/tokens").
		AddBasicAuth(user.Name)
	MakeRequest(t, req, http.StatusOK)

	// non-admin can get tokens for himself
	user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	req = NewRequest(t, "GET", "/api/v1/users/user2/tokens").
		AddBasicAuth(user.Name)
	MakeRequest(t, req, http.StatusOK)

	// non-admin can't get tokens for other users
	user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4})
	req = NewRequest(t, "GET", "/api/v1/users/user2/tokens").
		AddBasicAuth(user.Name)
	MakeRequest(t, req, http.StatusForbidden)
}

// TestAPIDeleteTokensPermission ensures that only the admin can delete tokens from other users
func TestAPIDeleteTokensPermission(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	admin := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user4 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4})

	// admin can delete tokens for other users
	createAPIAccessTokenWithoutCleanUp(t, "test-key-1", user2, []auth_model.AccessTokenScope{auth_model.AccessTokenScopeAll})
	req := NewRequest(t, "DELETE", "/api/v1/users/"+user2.LoginName+"/tokens/test-key-1").
		AddBasicAuth(admin.Name)
	MakeRequest(t, req, http.StatusNoContent)

	// non-admin can delete tokens for himself
	createAPIAccessTokenWithoutCleanUp(t, "test-key-2", user2, []auth_model.AccessTokenScope{auth_model.AccessTokenScopeAll})
	req = NewRequest(t, "DELETE", "/api/v1/users/"+user2.LoginName+"/tokens/test-key-2").
		AddBasicAuth(user2.Name)
	MakeRequest(t, req, http.StatusNoContent)

	// non-admin can't delete tokens for other users
	createAPIAccessTokenWithoutCleanUp(t, "test-key-3", user2, []auth_model.AccessTokenScope{auth_model.AccessTokenScopeAll})
	req = NewRequest(t, "DELETE", "/api/v1/users/"+user2.LoginName+"/tokens/test-key-3").
		AddBasicAuth(user4.Name)
	MakeRequest(t, req, http.StatusForbidden)
}

type permission struct {
	category auth_model.AccessTokenScopeCategory
	level    auth_model.AccessTokenScopeLevel
}

type requiredScopeTestCase struct {
	url                 string
	method              string
	requiredPermissions []permission
}

func (c *requiredScopeTestCase) Name() string {
	return fmt.Sprintf("%v %v", c.method, c.url)
}

// TestAPIDeniesPermissionBasedOnTokenScope tests that API routes forbid access
// when the correct token scope is not included.
func TestAPIDeniesPermissionBasedOnTokenScope(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// We'll assert that each endpoint, when fetched with a token with all
	// scopes *except* the ones specified, a forbidden status code is returned.
	//
	// This is to protect against endpoints having their access check copied
	// from other endpoints and not updated.
	//
	// Test cases are in alphabetical order by URL.
	testCases := []requiredScopeTestCase{
		{
			"/api/v1/admin/emails",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/admin/users",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/admin/users",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/admin/users/user2",
			"PATCH",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/admin/users/user2/orgs",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/admin/users/user2/orgs",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/admin/orgs",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryAdmin,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/notifications",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryNotification,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/notifications",
			"PUT",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryNotification,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/org/org1/repos",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryOrganization,
					auth_model.Write,
				},
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/packages/user1/type/name/1",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryPackage,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/packages/user1/type/name/1",
			"DELETE",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryPackage,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1",
			"PATCH",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1",
			"DELETE",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/branches",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/archive/foo",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/issues",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryIssue,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/media/foo",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/raw/foo",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/teams",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/teams/team1",
			"PUT",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/repos/user1/repo1/transfer",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Write,
				},
			},
		},
		// Private repo
		{
			"/api/v1/repos/user2/repo2",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		// Private repo
		{
			"/api/v1/repos/user2/repo2",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryRepository,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/user",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/user/emails",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/user/emails",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/user/emails",
			"DELETE",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/user/applications/oauth2",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
		{
			"/api/v1/user/applications/oauth2",
			"POST",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Write,
				},
			},
		},
		{
			"/api/v1/users/search",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
		// Private user
		{
			"/api/v1/users/user31",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
		// Private user
		{
			"/api/v1/users/user31/gpg_keys",
			"GET",
			[]permission{
				{
					auth_model.AccessTokenScopeCategoryUser,
					auth_model.Read,
				},
			},
		},
	}

	// User needs to be admin so that we can verify that tokens without admin
	// scopes correctly deny access.
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	assert.True(t, user.IsAdmin, "User needs to be admin")

	for _, testCase := range testCases {
		runTestCase(t, &testCase, user)
	}
}

// runTestCase Helper function to run a single test case.
func runTestCase(t *testing.T, testCase *requiredScopeTestCase, user *user_model.User) {
	t.Run(testCase.Name(), func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a token with all scopes NOT required by the endpoint.
		var unauthorizedScopes []auth_model.AccessTokenScope
		for _, category := range auth_model.AllAccessTokenScopeCategories {
			// For permissions, Write > Read > NoAccess.  So we need to
			// find the minimum required, and only grant permission up to but
			// not including the minimum required.
			minRequiredLevel := auth_model.Write
			categoryIsRequired := false
			for _, requiredPermission := range testCase.requiredPermissions {
				if requiredPermission.category != category {
					continue
				}
				categoryIsRequired = true
				if requiredPermission.level < minRequiredLevel {
					minRequiredLevel = requiredPermission.level
				}
			}
			unauthorizedLevel := auth_model.Write
			if categoryIsRequired {
				if minRequiredLevel == auth_model.Read {
					unauthorizedLevel = auth_model.NoAccess
				} else if minRequiredLevel == auth_model.Write {
					unauthorizedLevel = auth_model.Read
				} else {
					assert.FailNow(t, "Invalid test case: Unknown access token scope level: %v", minRequiredLevel)
				}
			}

			if unauthorizedLevel == auth_model.NoAccess {
				continue
			}
			cateogoryUnauthorizedScopes := auth_model.GetRequiredScopes(
				unauthorizedLevel,
				category)
			unauthorizedScopes = append(unauthorizedScopes, cateogoryUnauthorizedScopes...)
		}

		accessToken := createAPIAccessTokenWithoutCleanUp(t, "test-token", user, unauthorizedScopes)
		defer deleteAPIAccessToken(t, accessToken, user)

		// Request the endpoint.  Verify that permission is denied.
		req := NewRequest(t, testCase.method, testCase.url).
			AddTokenAuth(accessToken.Token)
		MakeRequest(t, req, http.StatusForbidden)
	})
}

// createAPIAccessTokenWithoutCleanUp Create an API access token and assert that
// creation succeeded.  The caller is responsible for deleting the token.
func createAPIAccessTokenWithoutCleanUp(t *testing.T, tokenName string, user *user_model.User, scopes []auth_model.AccessTokenScope) api.AccessToken {
	payload := map[string]any{
		"name":   tokenName,
		"scopes": scopes,
	}

	log.Debug("Requesting creation of token with scopes: %v", scopes)
	req := NewRequestWithJSON(t, "POST", "/api/v1/users/"+user.LoginName+"/tokens", payload).
		AddBasicAuth(user.Name)
	resp := MakeRequest(t, req, http.StatusCreated)

	var newAccessToken api.AccessToken
	DecodeJSON(t, resp, &newAccessToken)
	unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{
		ID:    newAccessToken.ID,
		Name:  newAccessToken.Name,
		Token: newAccessToken.Token,
		UID:   user.ID,
	})

	return newAccessToken
}

// deleteAPIAccessToken deletes an API access token and assert that deletion succeeded.
func deleteAPIAccessToken(t *testing.T, accessToken api.AccessToken, user *user_model.User) {
	req := NewRequestf(t, "DELETE", "/api/v1/users/"+user.LoginName+"/tokens/%d", accessToken.ID).
		AddBasicAuth(user.Name)
	MakeRequest(t, req, http.StatusNoContent)

	unittest.AssertNotExistsBean(t, &auth_model.AccessToken{ID: accessToken.ID})
}

// TestAPIPublicOnlyTokenRepos tests that token with public-only scope only shows public repos
func TestAPIPublicOnlyTokenRepos(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	publicOnlyScopes := []auth_model.AccessTokenScope{"public-only", "read:user", "read:repository"}
	publicOnlyAccessToken := createAPIAccessTokenWithoutCleanUp(t, "public-only-token", user, publicOnlyScopes)
	publicOnlyReq := NewRequest(t, "GET", "/api/v1/user/repos").
		AddTokenAuth(publicOnlyAccessToken.Token)
	publicOnlyResp := MakeRequest(t, publicOnlyReq, http.StatusOK)
	var publicOnlyRepos []api.Repository
	DecodeJSON(t, publicOnlyResp, &publicOnlyRepos)
	publicOnlyReposCaptured := make(map[string]int64)
	for _, repo := range publicOnlyRepos {
		publicOnlyReposCaptured[repo.Name] = repo.ID
		assert.False(t, repo.Private)
	}
	publicOnlyReposExpected := map[string]int64{
		"commits_search_test": 36,
		"commitsonpr":         58,
		"git_hooks_test":      37,
		"glob":                42,
		"repo-release":        57,
		"repo1":               1,
		"repo21":              32,
		"utf8":                33,
	}
	assert.Equal(t, publicOnlyReposExpected, publicOnlyReposCaptured)

	noPublicOnlyScopes := []auth_model.AccessTokenScope{"read:user", "read:repository"}
	noPublicOnlyAccessToken := createAPIAccessTokenWithoutCleanUp(t, "no-public-only-token", user, noPublicOnlyScopes)
	noPublicOnlyReq := NewRequest(t, "GET", "/api/v1/user/repos").
		AddTokenAuth(noPublicOnlyAccessToken.Token)
	noPublicOnlyResp := MakeRequest(t, noPublicOnlyReq, http.StatusOK)
	var allRepos []api.Repository
	DecodeJSON(t, noPublicOnlyResp, &allRepos)
	allPrivateReposCaptured := make(map[string]int64)
	for _, repo := range allRepos {
		if repo.Private {
			allPrivateReposCaptured[repo.Name] = repo.ID
		}
	}
	allPrivateReposExpected := map[string]int64{
		"big_test_private_4": 24,
		"lfs":                54,
		"readme-test":        56,
		"repo15":             15,
		"repo16":             16,
		"repo2":              2,
		"repo20":             31,
		"repo3":              3,
		"repo5":              5,
		"scoped_label":       55,
		"test_commit_revert": 59,
	}

	assert.Equal(t, allPrivateReposExpected, allPrivateReposCaptured)

	privateRepo2Req := NewRequest(t, "GET", "/api/v1/repos/user2/repo2").
		AddTokenAuth(noPublicOnlyAccessToken.Token)
	privateRepo2Resp := MakeRequest(t, privateRepo2Req, http.StatusOK)
	var repo2 api.Repository
	DecodeJSON(t, privateRepo2Resp, &repo2)
	assert.True(t, repo2.Private)

	publicOnlyPrivateRepo2Req := NewRequest(t, "GET", "/api/v1/repos/user2/repo2").
		AddTokenAuth(publicOnlyAccessToken.Token)
	MakeRequest(t, publicOnlyPrivateRepo2Req, http.StatusNotFound)

	// search query = repo1
	searchPublicOnlyReposReq := NewRequest(t, "GET", "/api/v1/repos/search?q=repo1").
		AddTokenAuth(publicOnlyAccessToken.Token)
	searchPublicOnlyReposResp := MakeRequest(t, searchPublicOnlyReposReq, http.StatusOK)
	var searchPublicOnlyRepos api.SearchResults
	DecodeJSON(t, searchPublicOnlyReposResp, &searchPublicOnlyRepos)
	var searchPublicOnlyRepoNamesCaptured []string
	for _, repo := range searchPublicOnlyRepos.Data {
		searchPublicOnlyRepoNamesCaptured = append(searchPublicOnlyRepoNamesCaptured, repo.Name)
	}
	searchPublicOnlyRepoNamesExpected := []string{
		"repo1",
		"repo10",
		"repo11",
	}

	assert.Equal(t, searchPublicOnlyRepoNamesExpected, searchPublicOnlyRepoNamesCaptured)

	notFoundPrivateByIDReq := NewRequest(t, "GET", "/api/v1/repositories/3").
		AddTokenAuth(publicOnlyAccessToken.Token)
	MakeRequest(t, notFoundPrivateByIDReq, http.StatusNotFound)

	foundPrivateByIDReq := NewRequest(t, "GET", "/api/v1/repositories/3").
		AddTokenAuth(noPublicOnlyAccessToken.Token)
	MakeRequest(t, foundPrivateByIDReq, http.StatusOK)
}

// TestAPIPublicOnlyTokenOrgs tests that token with public-only scope only shows public organizations
func TestAPIPublicOnlyTokenOrgs(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	session := loginUser(t, user.Name)

	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)
	org := api.EditOrgOption{
		Visibility: "private",
	}
	req := NewRequestWithJSON(t, "PATCH", "/api/v1/orgs/org3", &org).
		AddTokenAuth(token)
	MakeRequest(t, req, http.StatusOK)

	publicOnlyScopes := []auth_model.AccessTokenScope{"public-only", "read:user", "read:organization", "read:repository"}
	publicOnlyAccessToken := createAPIAccessTokenWithoutCleanUp(t, "public-only-token", user, publicOnlyScopes)
	publicOnlyReq := NewRequest(t, "GET", "/api/v1/user/orgs").
		AddTokenAuth(publicOnlyAccessToken.Token)
	publicOnlyResp := MakeRequest(t, publicOnlyReq, http.StatusOK)
	var publicOnlyOrgs []api.Organization
	DecodeJSON(t, publicOnlyResp, &publicOnlyOrgs)
	for _, org := range publicOnlyOrgs {
		assert.Equal(t, "public", org.Visibility)
	}
	noPublicOnlyScopes := []auth_model.AccessTokenScope{"read:user", "read:organization", "read:repository"}
	noPublicOnlyAccessToken := createAPIAccessTokenWithoutCleanUp(t, "no-public-only-token", user, noPublicOnlyScopes)
	noPublicOnlyReq := NewRequest(t, "GET", "/api/v1/user/orgs").
		AddTokenAuth(noPublicOnlyAccessToken.Token)
	noPublicOnlyResp := MakeRequest(t, noPublicOnlyReq, http.StatusOK)
	var allOrgs []api.Organization
	DecodeJSON(t, noPublicOnlyResp, &allOrgs)
	allOrgsCaptured := make(map[string]string)
	for _, org := range allOrgs {
		allOrgsCaptured[org.Name] = org.Visibility
	}
	allOrgsExpected := map[string]string{
		"org17":         "public",
		"org3":          "private",
		"private_org35": "private",
	}
	assert.Equal(t, allOrgsExpected, allOrgsCaptured)

	publicOnlySlashOrgReq := NewRequest(t, "GET", "/api/v1/orgs").
		AddTokenAuth(publicOnlyAccessToken.Token)
	publicOnlySlashOrgResp := MakeRequest(t, publicOnlySlashOrgReq, http.StatusOK)
	DecodeJSON(t, publicOnlySlashOrgResp, &publicOnlyOrgs)
	publicOrgsCaptured := make(map[string]string)
	for _, org := range publicOnlyOrgs {
		publicOrgsCaptured[org.Name] = org.Visibility
	}
	publicOrgsExpected := map[string]string{
		"org17": "public",
		"org19": "public",
		"org25": "public",
		"org26": "public",
		"org41": "public",
		"org6":  "public",
		"org7":  "public",
	}

	assert.Equal(t, publicOrgsExpected, publicOrgsCaptured)

	orgReposReq := NewRequest(t, "GET", "/api/v1/orgs/org17/repos").
		AddTokenAuth(noPublicOnlyAccessToken.Token)
	orgReposResp := MakeRequest(t, orgReposReq, http.StatusOK)
	var allOrgRepos []api.Repository
	DecodeJSON(t, orgReposResp, &allOrgRepos)
	allOrgReposCaptured := make(map[string]bool)
	for _, repo := range allOrgRepos {
		allOrgReposCaptured[repo.Name] = repo.Private
	}
	allOrgReposExpected := map[string]bool{
		"big_test_private_4": true,
		"big_test_public_4":  false,
	}
	assert.Equal(t, allOrgReposExpected, allOrgReposCaptured)

	publicOrgReposReq := NewRequest(t, "GET", "/api/v1/orgs/org17/repos").
		AddTokenAuth(publicOnlyAccessToken.Token)
	publicOrgReposResp := MakeRequest(t, publicOrgReposReq, http.StatusOK)
	var publicOrgRepos []api.Repository
	DecodeJSON(t, publicOrgReposResp, &publicOrgRepos)
	publicOrgReposCaptured := make(map[string]bool)
	for _, repo := range publicOrgRepos {
		publicOrgReposCaptured[repo.Name] = repo.Private
	}
	publicOrgReposExpected := map[string]bool{
		"big_test_public_4": false,
	}
	assert.Equal(t, publicOrgReposExpected, publicOrgReposCaptured)
}
