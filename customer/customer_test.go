package customer

import (
	"fmt"
	"testing"
	"time"

	"github.com/foomo/shop/utils"
)

func TestCustomerGetFormattedTime(t *testing.T) {
	now := time.Now()
	fmt.Println(utils.GetFormattedTime(now))
}
