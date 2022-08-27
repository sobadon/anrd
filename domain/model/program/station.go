package program

type Station string

const (
	StationOnsen = Station("onsen")
)

func (s Station) String() string {
	return string(s)
}
