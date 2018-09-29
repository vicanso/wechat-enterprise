package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
	"github.com/mozillazg/request"
	"github.com/vicanso/wechat-enterprise/config"
	"go.uber.org/zap"
)

var (
	logger         *zap.Logger
	corpID         string
	corpSecret     string
	corpAgentID    string
	getTokenURL    string
	sendMessageURL string
	validateToken  string
	fetchLock      = new(sync.Mutex)
	accessToken    *AccessToken
)

const (
	defaultTimeout = 10 * time.Second
)

type (
	// NoticeParams notic params
	NoticeParams struct {
		Users   string `valid:"runelength(1|1000)" json:"users,omitempty"`
		Type    string `valid:"in(text)" json:"type,omitempty"`
		Content string `valid:"runelength(1|2000)" json:"content,omitempty"`
		Token   string `json:"token,omitempty" valid:"runelength(10|30)"`
	}
	// GetTokenRes the response of get token
	GetTokenRes struct {
		AccessToken string `json:"access_token,omitempty"`
		ExpiresIn   int    `json:"expires_in,omitempty"`
	}
	// AccessToken the access token
	AccessToken struct {
		Value   string
		Expired int64
	}
	// SendMessageRes the response of send message
	SendMessageRes struct {
		ErrCode     int    `json:"errcode,omitempty"`
		ErrMsg      string `json:"errmsg,omitempty"`
		InvalidUser string `json:"invaliduser,omitempty"`
	}
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	c := zap.NewProductionConfig()
	l, err := c.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}
	logger = l
	corpID = config.GetString("corp.id")
	corpSecret = config.GetString("corp.secret")
	corpAgentID = config.GetString("corp.agentID")
	getTokenURL = config.GetString("corp.getToken")
	sendMessageURL = config.GetString("corp.sendMessage")
	validateToken = config.GetString("token")
	if corpID == "" ||
		corpSecret == "" ||
		corpAgentID == "" ||
		getTokenURL == "" ||
		sendMessageURL == "" {
		panic("id, secret, agent and token url can not be nil")
	}

}

func resJSON(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// resBadRequest 400出错响应（参数错误等）
func resBadRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	resJSON(w, []byte(`{
		"message": "`+err.Error()+`"
	}`))
}

func getHTTPClient() *http.Client {
	t := config.GetDurationDefault("timeout", defaultTimeout)

	return &http.Client{
		Timeout: t,
	}
}

// fetchAccessToken 获取访问的token
func fetchAccessToken() (token string, ttl int, err error) {
	url := fmt.Sprintf(getTokenURL, corpID, corpSecret)
	c := getHTTPClient()
	req := request.NewRequest(c)
	resp, err := req.Get(url)
	if err != nil {
		return
	}
	data, err := resp.Content()
	defer resp.Body.Close()

	if len(data) == 0 {
		err = errors.New("the response is nil")
		return
	}
	tokenInfo := GetTokenRes{}
	err = json.Unmarshal(data, &tokenInfo)
	if err != nil {
		return
	}
	token = tokenInfo.AccessToken
	ttl = tokenInfo.ExpiresIn
	return
}

func getToken() (token string) {
	now := time.Now().Unix()
	if accessToken != nil && accessToken.Expired > now {
		token = accessToken.Value
		return
	}
	fetchLock.Lock()
	defer fetchLock.Unlock()
	// 二次检查，如果有其它并发已更新token则不直接返回
	if accessToken != nil && accessToken.Expired > now {
		token = accessToken.Value
		return
	}

	value, ttl, err := fetchAccessToken()
	if err != nil {
		logger.Error("get access token fail",
			zap.Error(err),
		)
		return
	}
	logger.Info("fetch acccess token succcess")
	accessToken = &AccessToken{
		Value:   value,
		Expired: now + int64(ttl),
	}
	token = value
	return
}

// ping health check
func ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("pong"))
}

// notice
func notice(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resBadRequest(w, err)
		return
	}
	params := NoticeParams{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		resBadRequest(w, err)
		return
	}
	_, err = govalidator.ValidateStruct(params)
	if err != nil {
		resBadRequest(w, err)
		return
	}
	if params.Token != validateToken {
		resBadRequest(w, errors.New("token is invalid"))
		return
	}
	token := getToken()
	if token == "" {
		err = errors.New("get token fail")
		resBadRequest(w, err)
		return
	}

	data := map[string]interface{}{
		"touser":  params.Users,
		"msgtype": params.Type,
		"agentid": corpAgentID,
		"text": map[string]string{
			"content": params.Content,
		},
	}
	url := fmt.Sprintf(sendMessageURL, token)
	c := getHTTPClient()

	req := request.NewRequest(c)
	req.Json = data
	resp, err := req.Post(url)
	if err != nil {
		resBadRequest(w, err)
		return
	}
	body, err = resp.Content()
	defer resp.Body.Close()
	if err != nil {
		resBadRequest(w, err)
		return
	}
	sendMsgRes := SendMessageRes{}
	err = json.Unmarshal(body, &sendMsgRes)
	if err != nil {
		resBadRequest(w, err)
		return
	}
	if sendMsgRes.ErrCode != 0 {
		err = errors.New(sendMsgRes.ErrMsg)
		resBadRequest(w, err)
		return
	}
	if sendMsgRes.InvalidUser != "" {
		err = errors.New("invalid user:" + sendMsgRes.InvalidUser)
		resBadRequest(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	port := config.GetString("port")
	router := httprouter.New()
	router.GET("/ping", ping)
	router.POST("/notice", notice)

	logger.Info("the server will listen on " + port)
	log.Fatal(http.ListenAndServe(port, router))
}
