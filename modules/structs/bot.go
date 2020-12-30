package structs

type BotTask struct {
	UUID         string
	Repo         string
	Type         string
	Ref          string
	CommitID     string
	Event        string
	Token        string // token for this task
	Grant        string // permissions for this task
	EventPayload string
	Content      string
	Created      string
	StartTime    string
	EndTime      string
	RemoteURL    string
}
