package program

type Status string

const (
	StatusScheduled = Status("scheduled")
	StatusRecording = Status("recording")
	StatusDone      = Status("done")
	StatusFailed    = Status("failed")
)

func (s Status) String() string {
	return string(s)
}
