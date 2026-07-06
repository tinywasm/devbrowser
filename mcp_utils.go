package devbrowser

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/model"
)

func EncodeSchema(f model.Encodable) string {
	var s string
	_ = json.Encode(f, &s)
	return s
}
