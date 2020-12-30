package convert

import (
	"code.gitea.io/gitea/models"
	api "code.gitea.io/gitea/modules/structs"
)

func ToBotTask(repo *models.Repository, t *models.BotTask) *api.BotTask {
	return &api.BotTask{
		UUID:         t.UUID,
		Repo:         repo.FullName(),
		RemoteURL:    repo.HTMLURL(),
		Type:         t.Type,
		Ref:          t.Ref,
		CommitID:     t.CommitID,
		Event:        t.Event,
		Token:        t.Token,
		Grant:        t.Grant,
		EventPayload: t.EventPayload,
		Content:      t.Content,
		Created:      t.Created.Format("2006-01-02 15:04:05"),
		StartTime:    t.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:      t.EndTime.Format("2006-01-02 15:04:05"),
	}
}
