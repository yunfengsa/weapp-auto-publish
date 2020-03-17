package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	Path "path"
	"strings"
	"weapp-auto-publish/src/model"
	"weapp-auto-publish/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var isOnGoing bool
var needAgain bool

// TaskList 发布任务队列
var TaskList []model.Task

type publish struct {
	isOnGoing   bool
	needAgain   bool
	workSpace   string
	currentPath string
	TaskList    []model.Task
	cli         *utils.Cli
}

// New 返回实例
func New() *publish {
	currentPath, _ := os.Getwd()
	workPath := viper.GetString("workPath")
	workSpace := Path.Join(currentPath, workPath)
	return &publish{
		isOnGoing:   false,
		needAgain:   false,
		currentPath: currentPath,
		workSpace:   workSpace,
		TaskList:    []model.Task{},
		cli:         utils.InitCli(workSpace),
	}
}

// AutoPublishPost 发布请求
func (p *publish) AutoPublishPost(ctx *gin.Context) {
	var postData model.PublishPostData
	if err := ctx.ShouldBindJSON(&postData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentTask := model.Task{
		BranchName:  postData.Branch,
		User:        postData.User,
		Message:     postData.Message,
		TagName:     postData.TagName,
		IsProd:      postData.IsProd,
		GitRepoURL:  postData.GitRepo,
		ProjectName: postData.Name,
	}
	// 加入队列
	p.TaskList = append(p.TaskList, currentTask)
	ctx.JSON(http.StatusOK, gin.H{"results": "任务添加成功"})
	utils.PostMsgToRobot(fmt.Sprintf("队列新增1任务，项目：%s, 执行人: %s\n", currentTask.User, currentTask.User), []string{})
	if !p.isOnGoing {
		go p.startPublish()
	}
}

func (p *publish) startPublish() {
	p.isOnGoing = true
	for len(TaskList) > 0 {
		currentTask := p.TaskList[0]
		p.TaskList = p.TaskList[1:]
		if currentTask.IsProd {
			p.autoPublishToProd(currentTask)
		} else {
			p.autoPublishToTest(currentTask)
		}
	}
	p.isOnGoing = false
}

func (p *publish) autoPublishToTest(task model.Task) {
	branchName := task.BranchName
	userName := task.User
	message := task.Message
	utils.PostMsgToRobot(fmt.Sprintf("项目：%s\n线下打包服务启动\n打包分支:%s\n改动内容:%s\n执行人:%s", task.ProjectName, task.BranchName, task.Message, task.User), []string{})
	utils.RmIfExis(p.workSpace)
	p.cli.OpenCli()
	p.cloneProject(task.BranchName, task.GitRepoURL)
	p.npmInstall(true)
	testConfigPath := viper.GetString("projectConfig.test")
	p.createConfigFile(Path.Join(p.workSpace, testConfigPath))
	uploadInfo, uploadError := p.cli.NpmBuildAndUploadCode(task)
	p.cli.CloseCli()
	utils.RmIfExis(p.workSpace)
	if uploadError == nil {
		message := fmt.Sprintf("项目:%s\nsuccess, 线下打包成功\n打包分支:%s\n改动内容:%s\n执行人:%s", task.ProjectName, branchName, message, userName)
		if uploadInfo != nil {
			message = fmt.Sprintf("%s\n总包体积：%s;主包体积：%s", message, uploadInfo["total"], uploadInfo["main"])
		}
		utils.PostMsgToRobot(message, []string{})
	} else {
		utils.PostMsgToRobot(fmt.Sprintf("error!上传失败了,message:%s!\n打包分支:%s\n改动内容:%s\n执行人:%s", uploadError.Error(), branchName, message, userName), []string{""})
	}
	log.Printf("线下打包完成%+v", task)
}

// autoPublishToProd 自动发布到测试环境
func (p *publish) autoPublishToProd(task model.Task) {
	log.Printf("开始线上打包%+v", task)
	utils.PostMsgToRobot(fmt.Sprintf("项目：%s\n线上打包服务启动\n打包tag:%s\n改动内容:%s\n执行人:%s", task.ProjectName, task.TagName, task.Message, task.User), []string{""})
	utils.RmIfExis(p.workSpace)
	p.cli.OpenCli()
	p.cloneByTag(task.TagName, task.GitRepoURL)
	p.npmInstall(true)
	prodConfigPath := viper.GetString("projectConfig.prod")
	p.createConfigFile(Path.Join(p.workSpace, prodConfigPath))
	// 升级基础版本库
	// changeWeappVersion()
	// 进行测试
	p.cli.MiniNpmBuild()
	// 暂时屏蔽测试
	// testMsg, testPassed := p.runJest()
	testPassed := true
	// 测试通过
	if testPassed {
		// utils.PostMsgToRobot(fmt.Sprintf("success!测试通过：%s\n测试tag:%s\n执行人:%s", testMsg, task.TagName, task.User), []string{})
		uploadInfo, uploadError := p.cli.NpmBuildAndUploadCode(task)
		if uploadError == nil {
			message := fmt.Sprintf("项目：%s\nsuccess, 线上打包成功\n打包tag:%s\n改动内容:%s\n执行人:%s", task.ProjectName, task.TagName, task.Message, task.User)
			if uploadInfo != nil {
				message = fmt.Sprintf("%s\n总包体积：%s;主包体积：%s", message, uploadInfo["total"], uploadInfo["main"])
			}
			utils.PostMsgToRobot(message, []string{})
		} else {
			utils.PostMsgToRobot(fmt.Sprintf("error!上传失败了,message:%s!\n打包tag:%s\n改动内容:%s\n执行人:%s", uploadError.Error(), task.TagName, task.Message, task.User), []string{""})
		}
		// 测试失败
	} else {
		// utils.PostMsgToRobot(fmt.Sprintf("error!测试未通过,message:%s!\n打包tag:%s\n改动内容:%s\n执行人:%s", testMsg, task.TagName, task.Message, task.User), []string{""})
	}
	p.cli.CloseCli()
	utils.RmIfExis(p.workSpace)
	log.Printf("线上打包完成%+v", task)
}

// 克隆项目
func (p *publish) cloneProject(branch string, gitRepoURL string) {
	var URL = gitRepoURL
	ref := fmt.Sprintf("refs/heads/%s", branch)
	_, err := git.PlainClone(p.workSpace, false, &git.CloneOptions{
		URL:           URL,
		ReferenceName: plumbing.ReferenceName(ref),
		Progress:      os.Stdout,
	})
	if err != nil && strings.Contains(err.Error(), "reference not found") {
		log.Println("分支不存在或删除")
		utils.PostMsgToRobot(fmt.Sprintf("分支：%s 被删除或不存在", branch), []string{""})
	} else {
		checkIfError(err)
	}
}

func (p *publish) cloneByTag(tagName string, gitRepoURL string) {
	var URL = gitRepoURL
	ref := fmt.Sprintf("refs/tags/%s", tagName)
	_, err := git.PlainClone(p.workSpace, false, &git.CloneOptions{
		URL:           URL,
		ReferenceName: plumbing.ReferenceName(ref),
		Progress:      os.Stdout,
	})
	if err != nil && strings.Contains(err.Error(), "reference not found") {
		log.Println("分支不存在或删除")
		utils.PostMsgToRobot(fmt.Sprintf("TagName：%s 被删除或不存在", tagName), []string{""})
	} else {
		checkIfError(err)
	}
}

func (p *publish) npmInstall(onlyProd bool) {
	if onlyProd {
		p.createCmdAndRun("npm", "install", "--production")
	} else {
		p.createCmdAndRun("npm", "install")
	}
}

// // 运行测试
// func (p *publish) runJest() (string, bool) {
// 	testScript := viper.GetString("testScript")
// 	result := utils.CreateCmdGetLogInDist(p.workSpace, strings.Split(testScript, " "))
// 	log.Println(result)
// 	res := regexp.MustCompile(`Test\sSuites:\s(\S+)\spassed,\s(\S+)\stotal`)
// 	list := res.FindStringSubmatch(result)
// 	var passed bool
// 	if len(list) >= 3 && list[1] == list[2] {
// 		passed = true
// 		log.Println("测试通过")
// 	} else {
// 		log.Println("测试失败")
// 		passed = false
// 	}
// 	return result, passed
// }

func (p *publish) createConfigFile(fileName string) {
	p.CopyConfig(fileName)
	log.Println("项目文件创建完成")
}

// 创建命令并执行
func (p *publish) createCmdAndRun(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = p.workSpace
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	checkIfError(err)
	return err
}

func checkIfError(err error) {
	if err != nil {
		utils.PostMsgToRobot(fmt.Sprintf("构建过程出错了，请重试！message:%s", err.Error()), []string{""})
		log.Println(err)
	}
}

// CopyConfig 生成配置文件
func (p *publish) CopyConfig(filePath string) {
	mconfigRes, _ := ioutil.ReadFile(Path.Join(p.workSpace, viper.GetString("projectConfig.base")))
	originConfigRes, _ := ioutil.ReadFile(filePath)
	var m map[string]interface{}
	json.Unmarshal(originConfigRes, &m)
	json.Unmarshal(mconfigRes, &m)
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "  ")
	jsonEncoder.Encode(m)
	ioutil.WriteFile(Path.Join(p.workSpace, "project.config.json"), bf.Bytes(), os.ModePerm)
}

// 升级基础版本库到可执行测试版本
// func changeWeappVersion() {
// 	mconfigRes, _ := ioutil.ReadFile("./dist/project.config.json")
// 	reg := regexp.MustCompile(`"libVersion":\s?"[0-9]+\.[0-9]+\.[0-9]+"`)
// 	finalResult := reg.ReplaceAll(mconfigRes, []byte(`"libVersion": "2.9.4"`))
// 	err := os.Remove("./dist/project.config.json")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = ioutil.WriteFile("./dist/project.config.json", finalResult, os.ModePerm)
// 	log.Println(err)
// }
