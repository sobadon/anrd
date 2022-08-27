package timeutil

import "time"

func LocationJST() *time.Location {
	return time.FixedZone("Asia/Tokyo", 9*60*60)
}
