// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package notification

import (
	"code.gitea.io/gitea/modules/notification"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/services/notification/action"
	"code.gitea.io/gitea/services/notification/indexer"
	"code.gitea.io/gitea/services/notification/mail"
	"code.gitea.io/gitea/services/notification/ui"
	"code.gitea.io/gitea/services/notification/webhook"
)

// NewContext registers notification handlers
func NewContext() {
	notification.RegisterNotifier(ui.NewNotifier())
	if setting.Service.EnableNotifyMail {
		notification.RegisterNotifier(mail.NewNotifier())
	}
	notification.RegisterNotifier(indexer.NewNotifier())
	notification.RegisterNotifier(webhook.NewNotifier())
	notification.RegisterNotifier(action.NewNotifier())
}
