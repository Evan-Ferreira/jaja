package jobs

type JobType = string

const (
	JobTypeDocx="agent:docx"
)

type JobStatus = string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed JobStatus = "failed"
)