FROM golang

COPY ./spec /go/src/github.com/heroku/runtime-university-server/spec
COPY ./server /go/src/github.com/heroku/runtime-university-server/server
COPY ./cmd /go/src/github.com/heroku/runtime-university-server/cmd
COPY ./testdata /go/src/github.com/heroku/runtime-university-server/testdata
COPY ./vendor /go/src/github.com/heroku/runtime-university-server/vendor

WORKDIR /go/src/github.com/heroku/runtime-university-server/

RUN go get ./...
RUN go build ./...

