#!/usr/bin/env bash

TEST_PATH=github.com/foomo/shop
CONTAINER=$(docker run --rm -d -it -P mongo)
MONGO_PORT=$(docker inspect ${CONTAINER} | grep HostPort | sed 's/.*\"\([0-9]*\)".*/\1/g')

export MONGO_URL="mongodb://127.0.0.1:${MONGO_PORT}/shop"
export MONGO_URL_PRODUCTS="mongodb://127.0.0.1:${MONGO_PORT}/products"

ERRORS=""
RES=0

go test -v ${TEST_PATH}/crypto      || ERRORS="${ERRORS} crypto tests failed"       RES=1
go test -v ${TEST_PATH}/examples    || ERRORS="${ERRORS} examples tests failed"     RES=1
go test -v ${TEST_PATH}/shop_error  || ERRORS="${ERRORS} shop_error tests failed"   RES=1
go test -v ${TEST_PATH}/state       || ERRORS="${ERRORS} state tests failed"        RES=1
go test -v ${TEST_PATH}/unique      || ERRORS="${ERRORS} unique tests failed"       RES=1
go test -v ${TEST_PATH}/order       || ERRORS="${ERRORS} order tests failed"        RES=1
go test -v ${TEST_PATH}/customer    || ERRORS="${ERRORS} customer tests failed"     RES=1
go test -v ${TEST_PATH}/watchlist   || ERRORS="${ERRORS} watchlist failed"          RES=1

echo ${ERRORS}

docker stop ${CONTAINER}

exit ${RES}