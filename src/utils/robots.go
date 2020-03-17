package utils

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

// PostMsgToRobot 发送消息到企业微信
func PostMsgToRobot(content string, mentionList []string) {
	if viper.GetBool(`debug`) {
		log.Printf("Service RUN on DEBUG mode\n %s", content)
		return
	}
	var robotMsg string
	var qiWechatURL = viper.GetString("qiWeChat")
	// 没有配置企业微信机器人
	if qiWechatURL == "" {
		return
	}
	if mentionList != nil && len(mentionList) > 0 {
		robotMsg = fmt.Sprintf(`{
			"msgtype": "text", 
			"text": {
				"content": "%s",
				"mentioned_mobile_list": %+q
			}}`, content, mentionList)
	} else {
		robotMsg = fmt.Sprintf(`{
			"msgtype": "text", 
			"text": {
				"content": "%s"
			}}`, content)
	}
	req, err := http.NewRequest("POST", qiWechatURL, strings.NewReader(robotMsg))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	CheckAndLog(err)
	defer resp.Body.Close()
}
