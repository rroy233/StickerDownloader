package utils

import "encoding/json"

func JsonEncode(data any) string {
	out, _ := json.Marshal(data)
	return string(out)
}
