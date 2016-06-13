SHELL = "/bin/bash"

TEST_PATH = github.com/foomo/shop


clean:
	rm -f customer/diff-*
cleardbs:
	clear
	make clean
	go test -run TestDropAllCollections $(TEST_PATH)/test_utils
test:
	clear
	make clean
	go test -run Test $(TEST_PATH)/crypto
	go test -run Test $(TEST_PATH)/customer
	go test -run Test $(TEST_PATH)/examples
	go test -run Test $(TEST_PATH)/order
	go test -run Test $(TEST_PATH)/state
	go test -run Test $(TEST_PATH)/unique
	go test -run Test $(TEST_PATH)/shop_error
testv:
	clear
	make clean
	go test -v -run Test $(TEST_PATH)/crypto
	go test -v -run Test $(TEST_PATH)/customer
	go test -v -run Test $(TEST_PATH)/examples
	go test -v -run Test $(TEST_PATH)/order
	go test -v -run Test $(TEST_PATH)/state
	go test -v -run Test $(TEST_PATH)/unique
	go test -v -run Test $(TEST_PATH)/shop_error

