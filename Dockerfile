FROM golang:1.10.3-alpine3.7

RUN apk add --no-cache tzdata

ADD . /go/src/github.com/kaskavalci/bist100-tiwtter-bot/
WORKDIR /go/src/github.com/kaskavalci/bist100-tiwtter-bot

RUN go build -o /bist100-twitter-bot

ENTRYPOINT ["/bist100-twitter-bot"]
