package main

import (
	"fmt"
	"weapp-auto-publish/src/handlers"
	"weapp-auto-publish/src/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "net/http/pprof"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	if viper.GetBool(`debug`) {
		fmt.Println("Service RUN on DEBUG mode")
	}
}

func main() {
	router := gin.Default()
	router.Use(middleware.Cors)
	router.POST("auto-publish", handlers.New().AutoPublishPost)
	port := fmt.Sprintf(":%s", viper.GetString("server.port"))
	router.Run(port)
}
