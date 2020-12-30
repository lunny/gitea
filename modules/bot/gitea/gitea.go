package gitea

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/bot/runner"
	"code.gitea.io/gitea/modules/git"
	"github.com/pkg/errors"
)

func init() {
	runner.RegisterWorkerType(new(GiteaRunner))
}

type GiteaRunner struct {
}

func (gw *GiteaRunner) Name() string {
	return "gitea"
}

func (gw *GithubRunner) Detect(commit *git.Commit, event models.HookEventType, ref string) (bool, string, error) {
	return false, "", errors.New("Not implemented")
}
