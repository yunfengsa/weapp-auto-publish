package utils

import (
	"errors"
	"fmt"
	"log"
	Path "path"
	"regexp"
	"runtime"
	"strings"
	"time"
	"weapp-auto-publish/src/model"

	"github.com/spf13/viper"
)

// Cli 命令工具
type Cli struct {
	cliPath   string
	workSpace string
}

// InitCli 初始化cli
func InitCli(workSpace string) *Cli {
	cliPath := viper.GetString("cliPath")
	return &Cli{
		cliPath:   cliPath,
		workSpace: workSpace,
	}
}

// RestartCli 重启cli
func (cli *Cli) RestartCli() {
	cli.CloseCli()
	cli.OpenCli()
}

// OpenCli 打开开发者工具
func (cli *Cli) OpenCli() {
	results := CreateCmdGetLog(cli.cliPath, "-o")
	log.Println(results)
}

// CloseCli 关闭开发者工具
func (cli *Cli) CloseCli() {
	results := CreateCmdGetLog(cli.cliPath, "--quit")
	// 暂停五秒，命令行退出需要三秒的延迟
	time.Sleep(time.Duration(5) * time.Second)
	log.Println(results)
	log.Println("cli 关闭")
}

// MiniNpmBuild mini npm 构建
func (cli *Cli) MiniNpmBuild() string {
	return CreateCmdGetLog(cli.cliPath, "--build-npm", cli.workSpace)
}

// NpmBuildAndUploadCode 构建+上传
func (cli *Cli) NpmBuildAndUploadCode(task model.Task) (map[string]interface{}, error) {
	var uploadResult string
	var npmBUildResult string
	var outPutFile string
	switch runtime.GOOS {
	case "windows":
		npmBUildResult = cli.MiniNpmBuild()
		if task.TagName != "" {
			isProdTag, _ := regexp.MatchString("v[0-9]+\\.[0-9]+\\.[0-9]+", task.TagName)
			outPutFile = Path.Join(cli.workSpace, "uploadInfo", task.ProjectName, "release", fmt.Sprintf("%s.json", task.TagName))
			CreateFile(outPutFile)
			if isProdTag {
				uploadResult = CreateCmdGetLog(cli.cliPath, "-u", fmt.Sprintf("%s@%s", task.TagName[1:], cli.workSpace), "--upload-desc", task.TagName, "--upload-info-output", outPutFile)
			} else if strings.HasPrefix(task.TagName, "test") {
				uploadResult = CreateCmdGetLog(cli.cliPath, "-u", fmt.Sprintf("%s@%s", "test", cli.workSpace), "--upload-desc", task.TagName, "--upload-info-output", outPutFile)
			} else {
				PostMsgToRobot("tag标签格式错误! example: v1.0.0", []string{})
			}
		} else {
			now := time.Now()
			_, month, day := now.Date()
			outPutFile = Path.Join(cli.workSpace, "uploadInfo", task.ProjectName, "test", fmt.Sprintf("%d.json", now.Unix()))
			CreateFile(outPutFile)
			testVersion := fmt.Sprintf("m:%d.d:%d.h:%d", month, day, now.Hour())
			uploadResult = CreateCmdGetLog(cli.cliPath, "-u", fmt.Sprintf("%s@%s", testVersion, cli.workSpace), "--upload-desc", "this is auto publish", "--upload-info-output", outPutFile)
		}
	default:
		// createCmdAndRun("npm", "run", "upload-code")
		log.Println("当前只支持windows平台")
	}
	if strings.Contains(npmBUildResult, "npm success") {
		log.Println("npm build success!")
	} else {
		log.Println("npm build error!")
		return nil, errors.New(npmBUildResult)
	}
	if strings.Contains(uploadResult, "upload success") {
		log.Println("upload success!")
		upLoadInfo := readInfoJSON(outPutFile)
		log.Println(upLoadInfo)
		if upLoadInfo != nil && upLoadInfo.Size.Total > 0 {
			var mainSize = 0
			for _, v := range upLoadInfo.Size.Packages {
				if v.Name == "main" {
					mainSize = v.Size
					break
				}
			}
			return map[string]interface{}{
				"total": fmt.Sprintf("%dM", upLoadInfo.Size.Total/1024),
				"main":  fmt.Sprintf("%dM", mainSize/1024),
			}, nil
		}
		return nil, nil
	}
	log.Println("upload fail!")
	return nil, errors.New(uploadResult)
}

func (cli *Cli) checkIfError(err error) {
	if err != nil {
		PostMsgToRobot(fmt.Sprintf("构建过程出错了，请重试！message:%s", err.Error()), []string{""})
		log.Println(err)
		RmIfExis(cli.workSpace)
		cli.CloseCli()
	}
}
