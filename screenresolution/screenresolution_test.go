package screenresolution

import (
	"fmt"
	"testing"
)

func TestExampleGetPrimary(t *testing.T) {
	fmt.Println(GetPrimary().String())
	// Output: 1024x768
}
