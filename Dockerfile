FROM alpine

EXPOSE 3011 

ADD ./wechat-enterprise /

RUN apk --no-cache add ca-certificates \
  && update-ca-certificates

CMD ["/wechat-enterprise"]