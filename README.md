# eventops

### [简介](#简介) | [快速开始](#快速开始) | [安装](#安装) | [文档](./doc.md) | [TODO](./todo.md)

# 简介
eventops 是一个基于事件驱动的流水线工具，其目的是为了事件处理者根据事件的发生自动化的处理重复性的任务。

eventops 目前有以下三个工具

## eventops
server 端, 内部有 uc register pipeline event dialer 等五类 api

## eoctl
用于和 server 进行连接操作, 一种 C/S 架构类似 kubectl 和 kubernetes

## client
当 server 端无法直接访问 actuator 时候用作类似 vpn 功能的反向连接通道

# 快速开始

## 使用 docker-compose 部署
1. `git clone https://github.com/kakj-go/eventops.git`
2. `cd eventops/example/hello-world`
3. 修改 `docker-compose.yml 中的 [ip]`
4. `docker-compose up -d`

## 使用命令行部署
1. 自行安装 `mysql`, 并创建 `eventops` 库
2. 执行 `https://github.com/kakj-go/eventops/blob/master/tools/initdb/migrations/eventops.sql` 的 sql
3. 从 [安装](#安装) 了解如何获取 `eventops` 命令
4. 在 `/etc/eventops` 下创建 `config.yaml` 文件 (文件配置参考 [config.yaml](#configyaml))
5. `./eventops` 或者 `./eventops  --configFile=/etc/eventops/config.yaml` 来启动服务

下面是一个基础的 `config.yaml` 配置
```yaml
debug: true

# 要根据宿主机 ip 来修改或者增加 dns 解析也行
callbackAddress: http://evemtops:8080

mysql:
  user: root
  password: 123456
  address: 192.168.0.109

# 根据用户需要配置
#minio:
#  server: http://192.168.0.109:9000
#  accessKeyId: nf9SdeSrq7R4ffct
#  secretAccessKey: fPnttv7iWo5MQytu1IJ6SpK39078ED52
```

## eoctl 使用
> 注意: 下面命令成功的前提是 127.0.0.1:8080 能访问 eventops

1. 从 [安装](#安装) 了解如何获取 `eoctl` 工具
2. 注册用户 `eoctl register -s=http://127.0.0.1:8080 -u=kakj -p=123456 -e=2357431193@qq.com`
3. 登录用户 `eoctl login -s=http://127.0.0.1:8080 -u=kakj -p=123456`

## 使用 eoctl 创建 触发器定义 流水线定义 执行器定义
1. `git clone https://github.com/kakj-go/eventops.git && cd eventops/example/hello-world`
2. 修改 `osActuator.yaml` 配置
3. `eoctl actuator apply -f osActuator.yaml`
4. `eoctl pipeline apply -f pipelineDefinition.yaml`
5. `eoctl trigger apply -f triggerDefinition.yaml`

## 模拟发送事件
1. `eoctl event send -f example/hello-world/event.yaml`

## 查看流水线执行列表和获取详情
在 `osActuator` 声明的机器用户目录下，可以查看各种信息

```shell
[root@localhost 337]# pwd
/root/pipelines/237/tasks/337
[root@localhost 337]# ls
exit.code  nohup.log  nohup.pid  nohup.sh  response.txt  run.sh

# exit.code 文件记录用户的命令执行的退出码，只有退出码为 0 时任务才是成功状态

# nohup.log 文件记录用户命令的执行日志，其中包含标准输出和标准错误

# run.sh 里面包含了用户 task 中的 command 命令, 用户 command 命令前后会根据 task 是否使用文件类型的值和是否有出参来动态生成 mc 命令和 curl 命令

# nohup.sh 作为 run.sh 的父进程，目的时为了得到 run.sh 的 pid 和将 run.sh 置为后台运行进程

# response.txt 记录了 run.sh 回调 curl 命令的返回值，回调地址就对应 callbackAddress 的值

# nohup.pid 文件记录 run.sh 的执行进程 id
```

最后也可以使用 `eocli runtime list` 和 `eocli runtime get --id=pipelineId` 查看任务或者 `pipeline` 的执行情况

# 安装

## 获取方式

### 自行打包

```shell
git clone https://github.com/kakj-go/eventops.git
cd eventops

make eocli-linux-amd64
make eventops-linux-amd64
make client-linux-amd64
```

### 从 github 下载
`release` 中有 3 种工具可以下载 `eventops` `eoctl` 和 `client`

## 使用

### eventops
`eventops` 默认使用 `/etc/eventops/config.yaml` 配置文件

可以用 `eventops --configFile=B:\workspace\golang\eventops\conf\config.yaml` 来声明配置文件的位置

#### config.yaml
> 注意: 如果需要使用文件类型的事件内容，文件类型的入参，文件类型的出参, 文件类型的上下文参数则需要配置 minio

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

# 如果需要使用文件类型的事件内容，文件类型的入参，文件类型的出参, 文件类型的上下文参数则需要配置
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

# 事件处理的一些并发配置
# 以下是默认值
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

# 用户和校验
# 以下是默认值
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

启动 client 之前得先创建对应的 tunnel actuator，然后再根据 actuator 中的　tunnel 信息启动 client

`client` 启动 `./client --connect=ws://eventopsIp:eventopsPort/api/dialer/connect --id=actuatorDefinition中的clientKey --token=actuatorDefinition中的clientToken --user=username`


