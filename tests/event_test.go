package tests

import (
	"testing"

	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/test_utils"
)

func TestEventsSaveShopEvent(t *testing.T) {
	test_utils.DropAllCollections()
	foo()
}

func foo() {
	info := &event_log.Info{}
	event_log.SaveShopEvent(event_log.ActionTest, info, nil, "bla bla bla")
	event_log.ReportShopEvents()
}
