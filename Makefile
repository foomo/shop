
mongo:
	docker run --rm -d -it -p 27017:27017 mongo

clean:
	rm -f customer/diff-*

test: clean
	./scripts/test.sh
