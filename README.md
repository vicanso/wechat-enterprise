# wechat enterprise

此功能主要用来微信企业号的消息推送 


## 相关配置

- `corp.id` 企业ID，在企业信息中可以查看
- `corp.secret`  应用密钥，在应用程序的管理页面可查看
- `corp.agentID` 应用AgentID，在应用程序的管理页面可查看
- `token` 用于调用校验

## 发送消息

```bash
curl -XPOST -d '{
	"users": "userId,userId",
	"type": "text",
	"content": "要发送的内容",
	"token": "设置的校验token"
}' 'http://localhost:8011/notice'
```

## docker

```bash
docker build -t vicanso/wechat-enterprise .
```

```bash
docker run -d --restart=always \
  -e GO_ENV=production \
  -e CONFIG=/configs \
  -v ~/wechat-enterprise/production.yml:/configs/production.yml \
  --name notice vicanso/wechat-enterprise
```