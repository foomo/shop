SHELL = "/bin/bash"

TEST_PATH = github.com/foomo/shop
# invoke a single test by setting go test -v $(TEST_PATH)/shop

clean:
	rm -f customer/diff-*

test: clean
	./scripts/test.sh

install-test-dependencies:
	go get -u github.com/ventu-io/go-shortid
	go get -u github.com/bwmarrin/snowflake
	go get -u github.com/sergi/go-diff/...
	go get -u github.com/nbutton23/zxcvbn-go
	go get -u github.com/davecgh/go-spew
