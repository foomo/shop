package history

import (
	"encoding/json"
	"time"

	"github.com/foomo/shop/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Version struct {
	Current        int
	Previous       int // Previous version number (relevant, for example, atfer rollbacks)
	LastModifiedAt time.Time
}

func NewVersion() *Version {
	return &Version{
		Current:        0,
		Previous:       0,
		LastModifiedAt: utils.TimeNow(),
	}
}

func (v *Version) Increment() {
	v.Previous = v.Current
	v.Current = v.Current + 1
	v.LastModifiedAt = utils.TimeNow()
}

func (v *Version) GetVersion() int {
	return v.Current
}
func (v *Version) GetFormattedTime() string {
	return utils.GetFormattedTime(v.LastModifiedAt)
}

// DiffVersions compares to structs and returns the result as html.
// The html can be displayed with utils.OpenInBrowser()
func DiffVersions(versionA interface{}, versionB interface{}) (string, error) {
	jsonA, err := json.MarshalIndent(versionA, "", "	")
	if err != nil {
		return "", err
	}
	jsonB, err := json.MarshalIndent(versionB, "", "	")
	if err != nil {
		return "", err
	}
	d := diffmatchpatch.New()
	diffs := d.DiffMain(string(jsonA), string(jsonB), false)

	html := d.DiffPrettyHtml(diffs)
	return html, nil
}
