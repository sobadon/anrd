package program

type Station string

const (
	StationOnsen = Station("onsen")
	StationAgqr  = Station("agqr")
)

func (s Station) String() string {
	return string(s)
}
