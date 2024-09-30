// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrantAdditionalScopes(t *testing.T) {
	tests := []struct {
		grantScopes    string
		expectedScopes string
	}{
		{"openid profile email", ""},
		{"openid profile email groups", ""},
		{"openid profile email all", "all"},
		{"openid profile email read:user all", "read:user,all"},
		{"openid profile email groups read:user", "read:user"},
		{"read:user read:repository", "read:user,read:repository"},
		{"read:user write:issue public-only", "read:user,write:issue,public-only"},
		{"openid profile email read:user", "read:user"},
		{"read:invalid_scope", ""},
		{"read:invalid_scope,write:scope_invalid,just-plain-wrong", ""},
	}

	for _, test := range tests {
		t.Run(test.grantScopes, func(t *testing.T) {
			result := GrantAdditionalScopes(test.grantScopes)
			assert.Equal(t, test.expectedScopes, result)
		})
	}
}
