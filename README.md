# wechat enterprise

此功能主要用来测试微信企业号的消息推送 

curl -XPOST -H 'X-Token:xxx' -d '{
	"users": "XieShuZhou",
	"type": "text",
	"content": "程序重启"
}' 'http://127.0.0.1:3011/notice'

## docker


```bash
docker run -d --restart=always \
  -e PID=xx \
  -e SECRET=xx \
  -e AGENT=xxx \
  -e ACCESS_TOKEN=xxx \
  --name=lightning wechat-enterprise 
```