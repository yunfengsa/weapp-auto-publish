golang实现的微信小程序自动发布服务，接收http请求并完成相应发布任务

> 适合作为学习go 以及 搭建小程序发布服务的参考。测试应用场景比较小，覆盖的通用能力有限，感兴趣可以自己修改

#### 运行
* 安装小程序开发者工具、登陆、开启端口监听
* 修改配置文件config.json，[配置说明](#%e9%85%8d%e7%bd%ae%e8%af%b4%e6%98%8econfigjson)
* go mod tidy && ./build.sh && 运行相关环境产物
* 向该服务发送post请求，[post 参数](#post-%e5%8f%82%e6%95%b0)

#### 配置说明(config.json)

```js
{
  "debug": false, // 生产环境修改为false
  "server": {
    "port": 5788 // 服务启动端口
  },
  "projectConfig": {
    "base": "scripts/baseConfig.json", // 小程序公共配置文件路径
    "test": "scripts/testConfig.json", // 小程序测试环境配置文件路径
    "prod": "scripts/prodConfig.json" // 小程序公共线上配置文件路径
  },
  "qiWeChat": "", // 当前仅支持企业微信通知，企业微信地址
  "workPath": "dist", // 项目空间
  "cliPath": "cli.bat" // 小程序开发者工具命令行调用地址， 参考:https://developers.weixin.qq.com/miniprogram/dev/devtools/cli.html#%E8%87%AA%E5%8A%A8%E9%A2%84%E8%A7%88
}


```

#### post 参数

|参数|类型|说明|
|---|---|---|
|branch|string|分支名|
|tagName|string|tag名称|
|user|string|用户名|
|message|string|发布备注信息|
|gitRepo|string|仓库token|
|name|string|项目名称|
|isProd|true/false|true：使用tag false 使用branch|