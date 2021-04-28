package lmsensors

import (
	"encoding/json"

	"github.com/mdlayher/lmsensors"
)

// Update queries Linux Monitoring Sensors (lmsensors) by traversing sysfs.
// We think the lmsensors library is relatively safe because we do not expect
// sysfs reads to block.
func Update() ([]byte, error) {
	encoded := make([]byte, 0)
	scanner := lmsensors.New()
	data, err := scanner.Scan()
	if err == nil {
		encoded, err = json.Marshal(data)
		if err != nil {
			return []byte(nil), err
		}
	}

	return encoded, err
}
