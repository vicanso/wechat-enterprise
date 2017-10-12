package main

import (
	"bytes"
	"errors"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Token struct {
	value   string
	expired int32
}

var PID = os.Getenv("PID")
var SECRET = os.Getenv("SECRET")
var AGENT = os.Getenv("AGENT")
var tokenInfo = Token{}
var ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")

func getToken() string {
	if len(tokenInfo.value) != 0 && tokenInfo.expired > int32(time.Now().Unix()) {
		return tokenInfo.value
	}
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + PID + "&corpsecret=" + SECRET
	c := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	token, err := jsonparser.GetString(body, "access_token")
	if err != nil {
		log.Fatal(err)
		return ""
	}
	expires, err := jsonparser.GetInt(body, "expires_in")
	if err != nil {
		log.Fatal(err)
		return ""
	}
	now := int32(time.Now().Unix())
	tokenInfo.value = token
	tokenInfo.expired = now + (int32(expires)) - (5 * 60)
	log.Printf("get token success, expired:%d", tokenInfo.expired)
	return token
}

func sendNotice(users, msgType, content string) (string, error) {
	token := getToken()
	if len(token) == 0 {
		return "", errors.New("Get token fail")
	}
	url := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + token
	jsonStr := `{
		"touser": "${users}",
		"msgtype": "${msgType}",
		"agentid": ${agent},
		"text": {
			"content": "${content}"
		}
	}`
	jsonStr = strings.Replace(jsonStr, "${users}", users, 1)
	jsonStr = strings.Replace(jsonStr, "${msgType}", msgType, 1)
	jsonStr = strings.Replace(jsonStr, "${content}", content, 1)
	jsonStr = strings.Replace(jsonStr, "${agent}", AGENT, 1)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	invalidUsers, err := jsonparser.GetString(body, "invaliduser")
	if err != nil {
		return "", err
	}
	return invalidUsers, nil
}

func noticeCaptchaServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	if req.Header.Get("X-Token") != ACCESS_TOKEN {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "token is invalid"}`))
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "the body is empty"}`))
		return
	}
	users, _ := jsonparser.GetString(body, "users")
	msgType, _ := jsonparser.GetString(body, "type")
	content, _ := jsonparser.GetString(body, "content")

	invalidUsers, err := sendNotice(users, msgType, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "send message fail"}`))
		return
	}
	resStr := strings.Replace(`{"invalidUsers": "${1}"}`, "${1}", invalidUsers, 1)
	w.Write([]byte(resStr))
	log.Printf("send notice to %s success", users)
}

func pingServe(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func main() {
	_ = getToken()
	http.HandleFunc("/ping", pingServe)
	http.HandleFunc("/notice", noticeCaptchaServe)
	log.Println("server is at :3011")
	if err := http.ListenAndServe(":3011", nil); err != nil {
		log.Fatal(err)
	}
}
