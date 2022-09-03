# eventops
基于事件驱动的 ops 工具或者说流水线.

### [简介](#简介) | [快速开始](#快速开始) | [安装](#安装) | [文档](#./doc.md) | [TODO](./todo.md)

# 简介
eventops 是一个基于事件驱动的流水线工具，其目的是为了事件处理者根据事件的发生自动化的处理重复性的任务。

eventops 目前有以下三个工具

## eventops
作为 server 端, 内部有 uc register pipeline event dialer 等五类 api

## eoctl
用于和 server 进行连接操作。一种 C/S 架构类似 kubectl 和 kubernetes

## client
在没有公网 ip 的 actuator 中启动，用于和 eventops server 建立 websocket 连接，从而 server 可以管理这种 actuator

# 快速开始

## 创建 mysql 库和表
1. `mysql` 安装自行参考网上教程，或者直接使用 `docker` 进行安装
2. `mysql` 安装完毕后，从 `https://github.com/kakj-go/eventops.git` 的 `migrations` 目录下获取 `eventops.sql` 去创建库和表

## 启动 eventops
1. 从 [安装](#安装) 了解如何获取 `eventops` 工具
2. 在某个目录下创建 `config.yaml` 文件 (文件配置参考 [config.yaml](###config.yaml))
3. 启动 `eventops --configFile=/etc/eventops/config.yaml`

## 使用 eoctl
> 注意下面的登录和注册的前提是 eventops 启动地址是 127.0.0.1:8080
1. 从 [安装](#安装) 了解如何获取 `eoctl` 工具
3. 注册用户 `eoctl register -s=http://127.0.0.1:8080 -u=kakj -p=123456 -e=2357431193@qq.com`
4. 登录用户 `eoctl login -s=http://127.0.0.1:8080 -u=kakj -p=123456`

## 使用 eoctl 创建 actuatorDefinition, pipelineDefinition, triggerDefinition
1. `git clone https://github.com/kakj-go/eventops.git && cd eventops`
2. 修改 `example/hello-world/osActuator.yaml` 配置,
3. `eoctl actuator apply -f example/hello-world/osActuator.yaml`
4. `eoctl pipeline apply -f example/hello-world/pipelineDefinition.yaml`
5. `eoctl trigger apply -f example/hello-world/triggerDefinition.yaml`

## 模拟发送事件
1. `eoctl event send -f example/hello-world/event.yaml`

## 查看流水线执行列表和获取详情
1. `eocli runtime list`
2. `eocli runtime get --id=pipelineId`

# 安装

## 获取方式

### 自行打包

```shell
git clone https://github.com/kakj-go/eventops.git
cd eventops
make xxx
```

### 从 github 下载
`release` 有 3 种工具可以下载 `eventops` `eoctl` 和 `client`

## 使用

### eventops
`eventops` 默认使用 `/etc/eventops/config.yaml` 配置文件

可以用 `eventops --configFile=B:\workspace\golang\eventops\conf\config.yaml` 来声明配置文件的位置

#### config.yaml
```yaml
# 启动的端口 (必填)
port: 8080
# 是否是 debug 模式启动 
debug: true 

# 任务回调的 eventops 地址 (必填)
# 该地址要 task 能访问的 eventops 地址
callbackAddress: http://127.0.0.1:8080

# mysql 连接地址 (必填)
mysql:
  user: root
  password: 123456
  address: 127.0.0.1
  port: 3306
  db: eventops

# 如果需要使用文件类型的 outputs inputs contexts，则需要配置
#minio:
#  # minio 地址
#  server: http://127.0.0.1:9000 
#  # minio 用户的 keyId
#  accessKeyId: nf9SdeSrq7R4ffct
#  # minio 用户的 accessKey
#  secretAccessKey: fPnttv7iWo5MQytu1IJ6SpK39078ED52
#  # 是否开启 ssl
#  ssl: false
#  # 基础 bucket
#  basePath: eventops

# 事件处理的一些并发配置 (必填)
event:
  process:
    # 事件处理的 buffer
    bufferSize: 500
    # 并发处理这些 buffer 的携程数
    workNum: 5
    # 对于 triggerDefinition 的缓存大小
    triggerCacheSize: 10000
    # 循环加载数据库中事件的间隔事件
    loopLoadEventInterval: 300
    # 事件 processing 超时事件
    processingOverTime: 120

# 用户和校验 (必填)
uc:
  # 登录的 token 过期时间 
  loginTokenExpiresTime: 315360000
  # token 的 Signature
  loginTokenSignature: MYSQL_SIGNATURE
  # 无需要验证登录的 api
  auth:
    whiteUrlList:
      - /api/user/register
      - /api/user/login
      - /api/dialer/connect
      - /api/pipeline/callback
```

### eoctl
`eventops` 的 `cli` 工具

`eoctl` 需要进行注册和登录，登录后会在 `homedir` 下创建 `.eoctl.yaml` 的认证文件

`eoctl register -s=http://eventopsAddress:eventopsPort -u=username -p=password -e=email`

`eoctl login -s=http://eventopsAddress:eventopsPort -u=username -p=password`

登录成功后就可以使用 `eoctl -h` 来操作 `eventops` 了

### client
`client` 作为 `eventops` 和 `actuator` 的连接通道，可以抽象成 `vpn`

如果你的 `eventops` 无法直接访问 `actuator` 的地址。那么 `client` 是一种反向连接的工具, `client` 启动会主动和 `eventops` 建立 `websocket` 连接，
然后 `eventops` 通过 `websocket` 连接通道对 `actuator` 进行管理

`client` 启动 `./client --id=actuatorDefinition中的clientKey --token=actuatorDefinition中的clientToken --user=username`


