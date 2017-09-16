const Koa = require('koa');
const mount = require('koa-mounting');
const koaLog = require('koa-log');
const request = require('superagent');
const bodyparser = require('koa-bodyparser');
const Joi = require('joi');

const config = require('./config');

let accessTokenInfo = null;

// 获取 access token
async function getAccessToken() {
  // 如果token未过期，直接返回
  if (accessTokenInfo && accessTokenInfo.validPeriod > Date.now()) {
    return accessTokenInfo.token;
  }
  const url = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=${config.pid}&corpsecret=${config.secret}`;
  const res = await request.get(url);
  if (res.body.errcode !== 0) {
    throw new Error(res.body.errmsg || 'unknown error');
  }
  const token = res.body.access_token;
  // 有效期提前5分钟失效
  const validPeriod = (Date.now() + (res.body.expires_in * 1000)) - (5 * 60 * 1000);
  accessTokenInfo = {
    token,
    validPeriod,
  };
  return token;
}

async function sendNotice(params) {
  const data = {
    touser: params.users,
    msgtype: params.type,
    agentid: config.agent,
    text: {
      content: params.content,
    },
  };
  const token = await getAccessToken();
  const url = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=${token}`;
  const res = await request.post(url, data);
  const {
    errcode,
    invaliduser,
    errmsg,
  } = res.body;
  if (errcode !== 0) {
    throw new Error(errmsg);
  }
  if (invaliduser) {
    console.error(`invalid user:${invaliduser}`);
  }
}


async function notice(ctx) {
  if (ctx.method !== 'POST') {
    throw new Error('method is not allow');
  }
  if (ctx.get('Token') !== 'k1nG2QY9ef') {
    throw new Error('token is invalid');
  }
  const {
    value,
    error,
  } = Joi.validate(ctx.request.body, {
    users: Joi.string().required(),
    type: Joi.string().default('text'),
    content: Joi.string().required(),
  });
  if (error) {
    throw new Error(error.message);
  }
  await sendNotice(value);
  ctx.status = 201;
}

const app = new Koa();

app.use((ctx, next) => next().catch((err) => {
  ctx.status = 500;
  ctx.body = {
    message: err.message,
  };
}));

app.use(mount('/ping', (ctx) => {
  ctx.body = 'pong';
}));

app.use(koaLog());

app.use(bodyparser());

app.use(mount('/notice', notice));

app.listen(3011);
console.info('The server is listen on 3011');
