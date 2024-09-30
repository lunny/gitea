// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package utils

import "code.gitea.io/gitea/services/context"

// check if api token contains `public-only` scope
func PublicOnlyToken(ctx *context.APIContext, scopeKey string) bool {
	publicScope, _ := ctx.Data[scopeKey].(bool)
	return publicScope
}
