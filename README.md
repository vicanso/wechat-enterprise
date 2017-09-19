# wechat enterprise

此功能主要用来测试微信企业号的消息推送 

## docker


```bash
docker run -d --restart=always \
  -e PID=xx \
  -e SECRET=xx \
  -e AGENT=xxx \
  -e ACCESS_TOKEN=xxx \
  --name=lightning wechat-enterprise 
```