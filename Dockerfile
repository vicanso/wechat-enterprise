FROM golang:1.11-alpine as builder

ADD ./ /go/src/github.com/vicanso/wechat-enterprise

RUN apk update \
  && apk add git \
  && go get -u github.com/golang/dep/cmd/dep \
  && cd /go/src/github.com/vicanso/wechat-enterprise \
  && dep ensure \
  && GOOS=linux GOARCH=amd64 go build -tags netgo -o wechat-enterprise

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/vicanso/wechat-enterprise/wechat-enterprise /usr/local/bin/wechat-enterprise
COPY --from=builder /go/src/github.com/vicanso/wechat-enterprise/configs /configs

CMD [ "wechat-enterprise" ]

