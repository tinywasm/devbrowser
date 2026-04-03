package devbrowser

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
)

func EncodeSchema(f fmt.Fielder) string {
	var s string
	_ = json.Encode(f, &s)
	return s
}
