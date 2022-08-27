package program

type StreamType string

const (
	StreamTypeBroadcast = StreamType("broadcast")
	StreamTypeOndemand  = StreamType("ondemand")
)

func (s StreamType) String() string {
	return string(s)
}
