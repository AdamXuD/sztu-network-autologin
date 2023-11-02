package main

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

var GETIP_URL = "http://www.qq.com/"
var TEST_URL = "https://www.baidu.com/"
var LOGIN_URL = "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/cmcc_login/"
var ONSUCCESS_URL = "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/success/"
var ONFAIL_URL = "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/fail/"
var CHECKSTATUS_URL = "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/cmcc_login_result/"

func isOnline() bool {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	resp, err := client.R().Get(TEST_URL)
	if err != nil {
		return false
	}
	return resp.StatusCode() == 200
}

func parseTokenFromHtml(html []byte) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return ""
	}
	token, exists := doc.Find("form input").First().Attr("value")
	if !exists {
		return ""
	}
	return token
}

func getNowIP() string {
	client := resty.New()
	resp, err := client.R().Get(GETIP_URL)
	if err != nil {
		return ""
	}
	location := resp.RawResponse.Request.URL.String()
	urlObj, _ := url.Parse(location)
	queryDict, _ := url.ParseQuery(urlObj.RawQuery)
	return queryDict.Get("wlanuserip")
}

func postLogin(usrip string) string {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Cookie", fmt.Sprintf("portal_usrname=%s", os.Getenv("USER_ID"))).
		SetFormData(map[string]string{
			"usrname":        os.Getenv("USER_ID"),
			"passwd":         os.Getenv("PASSWORD"),
			"treaty":         "on",
			"nasid":          fmt.Sprint(3),
			"usrmac":         os.Getenv("DEVICE_MAC"),
			"usrip":          usrip,
			"basip":          "172.17.127.254",
			"success":        ONSUCCESS_URL,
			"fail":           ONFAIL_URL,
			"offline":        fmt.Sprint(0),
			"portal_version": fmt.Sprint(1),
			"portal_papchap": "pap",
		}).Post(LOGIN_URL)
	if err != nil {
		return ""
	}
	if resp.StatusCode() == 200 {
		return parseTokenFromHtml(resp.Body())
	} else {
		return ""
	}
}

func postToken(token string) bool {
	client := resty.New()
	resp, err := client.R().SetFormData(map[string]string{
		"cmcc_login_value": token,
	}).SetHeader("Cookie", fmt.Sprintf("portal_usrname=%s", os.Getenv("USER_ID"))).Post(LOGIN_URL)
	if err != nil {
		return false
	}
	return resp.StatusCode() == 200
}

func getStatus(token string) bool {
	client := resty.New()
	for i := 0; i < 5; i++ {
		resp, err := client.R().SetFormData(map[string]string{
			"l": token,
		}).SetHeader("Cookie", fmt.Sprintf("portal_usrname=%s", os.Getenv("USER_ID"))).Post(CHECKSTATUS_URL)
		if err != nil {
			return false
		}
		if resp.StatusCode() == 200 && resp.String() == "success" {
			return true
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	return false
}

func checkStatus() bool {
	client := resty.New()
	resp, err := client.R().Get(ONSUCCESS_URL)
	if err != nil {
		return false
	}
	return resp.StatusCode() == 200
}

func login() (bool, string) {
	// usrip := getNowIP()
	// if usrip == "" {
	// 	return false, "无法获取用户IP。"
	// }

	usrip := "10.119.4.28"

	token := postLogin(usrip)
	if token == "" {
		return false, "无法正确获取Token。"
	}

	if !postToken(token) {
		return false, "无法提交登录状态。"
	}

	if !getStatus(token) {
		return false, "登录状态失败。"
	}

	if !checkStatus() {
		return false, "登录失败。"
	}

	return true, "登录成功."
}

func printLog(logStr string) {
	log.Printf("%s | %s", time.Now().Format("2006-01-02 15:04:05"), logStr)
}

func mainLoop() {
	checkInterval, err := strconv.ParseInt(os.Getenv("CHECK_INTERVAL"), 10, 64)
	if err != nil {
		checkInterval = 5
	}
	retryMaxCount, err := strconv.ParseInt(os.Getenv("RETRY_MAXCOUNT"), 10, 64)
	if err != nil {
		retryMaxCount = 5
	}
	printLog("网络离线检测已启动。")
	count := 0
	for count < int(retryMaxCount) {
		for isOnline() {
			time.Sleep(time.Duration(checkInterval) * time.Second)
		}
		printLog(fmt.Sprintf("网络离线，正在进行第%d次重连... ...", count+1))
		res, hint := login()
		if res {
			count = 0
		} else {
			count++
		}
		printLog(hint)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		printLog("未加载.env文件到环境变量中。")
	} else {
		printLog("已加载.env文件到环境变量中。")
	}
	mainLoop()
}
