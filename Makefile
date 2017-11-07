SHELL = "/bin/bash"

TEST_PATH = github.com/foomo/shop


clean:
	rm -f customer/diff-*

clear-dbs: clean
	go test -run TestDropAllCollections $(TEST_PATH)/test_utils

test: clean
	$(eval CONTAINER := $(shell docker run --rm -d -it -p "27017:27017" mongo))
	MONGO_URL="mongodb://localhost/shop" go test  ./... || echo "Tests failed"
	docker stop $(CONTAINER)

install-test-dependencies:
	go get -u github.com/ventu-io/go-shortid
	go get -u github.com/bwmarrin/snowflake
	go get -u github.com/sergi/go-diff/...
	go get -u github.com/nbutton23/zxcvbn-go
	go get -u github.com/davecgh/go-spew
