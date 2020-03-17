package utils

import "log"

// CheckAndLog 检查并输出
func CheckAndLog(err error) {
	if err != nil {
		log.Println(err)
	}
}
