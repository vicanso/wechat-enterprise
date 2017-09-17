FROM alpine

EXPOSE 3011 

ADD ./wechat-enterprise /

CMD ["/wechat-enterprise"]