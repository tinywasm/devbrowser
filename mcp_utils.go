package devbrowser

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
)

func EncodeSchema(f fmt.Encodable) string {
	var s string
	_ = json.Encode(f, &s)
	return s
}
