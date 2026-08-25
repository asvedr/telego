//go:build stdjson

package json

import "encoding/json"

func init() {
	Marshal = json.Marshal
	Unmarshal = json.Unmarshal
}
