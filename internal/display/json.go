package display

import (
	"encoding/json"
)

func DumpJson(obj any) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		b, _ := json.MarshalIndent(err, "", "  ")
		return string(b)
	} else {
		return string(b)
	}
}
