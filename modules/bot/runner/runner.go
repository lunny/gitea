package runner

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"
)

var (
	runnerTypes = make(map[string]RunnerType)
)

type RunnerType interface {
	Name() string
	Detect(commit *git.Commit, event models.HookEventType, ref string) (bool, string, error)
}

func RegisterRunnerType(runnerType RunnerType) {
	runnerTypes[runnerType.Name()] = runnerType
}

func GetRunnerType(name string) RunnerType {
	return runnerTypes[name]
}

func GetRunnerTypes() map[string]RunnerType {
	return runnerTypes
}
