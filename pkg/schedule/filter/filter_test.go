package filter

import "testing"

func TestMatchModel(t *testing.T) {
	println(matchModel(`.*1080\sTi`, "GeForce GTX 1080 Ti"))
}
