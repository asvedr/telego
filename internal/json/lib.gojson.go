//go:build !stdjson

package json

import json "github.com/asvedr/go-json"

func init() {
	Marshal = json.Marshal
	Unmarshal = json.Unmarshal
}
