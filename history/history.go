package history

import (
	"time"

	"github.com/foomo/shop/utils"
)

type Version struct {
	Number    int
	TimeStamp time.Time
}

func (v *Version) Increment() {
	v.Number = v.Number + 1
	v.TimeStamp = time.Now()
}

func (v *Version) GetVersion() int {
	return v.Number
}
func (v *Version) GetFormattedTime() string {
	return utils.GetFormattedTime(v.TimeStamp)
}
