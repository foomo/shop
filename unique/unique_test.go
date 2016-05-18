package unique

import (
	"fmt"
	"testing"
)

func TestCreateUniqueIds(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(GetNewID())
	}
}
