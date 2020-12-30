package bot

import (
	_ "code.gitea.io/gitea/modules/bot/github" // register worker type
)

func Init() error {
	return nil
}
