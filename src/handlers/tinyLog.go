package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"weapp-auto-publish/src/model"

	"github.com/gin-gonic/gin"
)

// TinyLog 返回小程序构建打包数据
func TinyLog(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
	}()
	result := readDataFromLocal()
	if result == nil {
		result = map[string]interface{}{
			"error": "error",
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"result": result,
	})
	// }()
}

func readDataFromLocal() interface{} {
	baseDir := "./uploadInfo"
	response := map[string]map[string]map[string]model.TinyLogo{}
	projectList, err := ioutil.ReadDir(baseDir)
	checkReadFileError(err)
	for _, project := range projectList {
		response[project.Name()] = map[string]map[string]model.TinyLogo{}
		envList, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", baseDir, project.Name()))
		checkReadFileError(err)
		for _, env := range envList {
			response[project.Name()][env.Name()] = map[string]model.TinyLogo{}
			infoList, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", baseDir, project.Name(), env.Name()))
			checkReadFileError(err)
			for _, info := range infoList {
				infoBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s/%s", baseDir, project.Name(), env.Name(), info.Name()))
				checkReadFileError(err)
				infoStruct := model.TinyLogo{}
				err = json.Unmarshal(infoBytes, &infoStruct)
				checkReadFileError(err)
				response[project.Name()][env.Name()][info.Name()] = infoStruct
			}
		}
	}
	return response
}

func checkReadFileError(err error) {
	if err != nil {
		log.Println(err)
	}
}
