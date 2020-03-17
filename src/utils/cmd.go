package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	Path "path"
	"time"

	model "weapp-auto-publish/src/model"
)

// IsExist 文件是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// Copy 拷贝文件
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer source.Close()
	log.Println(dst)
	destination, err := os.Create(dst)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// CreateFile 创建文件
func CreateFile(fullPath string) {
	path, fileName := Path.Split(fullPath)
	if !IsExist(path) {
		os.MkdirAll(path, os.ModePerm)
	}

	filePath := Path.Join(path, fileName)
	if !IsExist(filePath) {
		os.Create(filePath)
	}
}

// CreateCmdGetLog 创建命令并执行
func CreateCmdGetLog(args ...string) string {
	cmd := exec.Command(args[0], args[1:]...)
	output, _ := cmd.CombinedOutput()
	return string(output)
}

// CreateCmdGetLogInDist 在工作区域执行命令
func CreateCmdGetLogInDist(workSpace string, args []string) string {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workSpace
	output, _ := cmd.CombinedOutput()
	return string(output)
}

func readInfoJSON(filePath string) *model.UpLoadInfo {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	upLoadInfo := model.UpLoadInfo{}
	if err == nil {
		err = json.Unmarshal(fileContent, &upLoadInfo)
	} else {
		log.Println(err)
	}
	if err == nil {
		return &upLoadInfo
	}
	return nil
}

// RmIfExis 清空工作空间文件
func RmIfExis(workSpace string) {
	_, err := os.Stat(workSpace)
	if err == nil {
		err := os.RemoveAll(workSpace)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Duration(2) * time.Second)
	} else {
		log.Println(err)
	}
}
