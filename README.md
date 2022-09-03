# eventops
基于事件驱动的 ops 工具或者说流水线.

### [简介](#简介) | [快速开始](#快速开始) | [安装](#安装) | [文档](#文档) | [TODO](./TODO.md)

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
2. 在某个目录下创建 `config.yaml` 文件 (文件配置参考 [eventops](###eventops))
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

### 从 github release 下载
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

# 文档
进行中

#### inputs
在 `triggerDefinition` 中的 `inputs` 就是声明执行 `pipeline` 的时候传递的入参

在 `pipelineDefinition` 中的 `inputs` 就是声明该流水线运行时需要那些入参

在 `pipeline` 类型的 `task` 中的 `inputs` 就是声明执行 `pipeline` 的时候传递的入参

`task` 可以使用 `${{ inputs.inputName }}` 来使用入参的值

#### outputs
在 `pipelineDefinition` 中的 `outputs` 就是声明该流水线被 `task` 调用的时候会产生那些出参

在 `task` 中的 `outputs` 就是声明该任务会有那些出参，非 `pipeline` 类型的任务出参可以声明环境变量的值或者文件，
`pipeline` 类型的任务只能引用那条流水线所具有的出参

`task` 可以使用 `${{ outputs.taskName.taskOutputName }}` 来使用值

#### contexts
在 `pipelineDefinition` 中声明全局变量，使用 `task` 的 `outputs` 字段 `setToContexts` 进行设值

`task` 可以使用 `${{ contexts.contextName }}` 来使用值

#### image
`pipeline` 类型的 `task image` 是对应 `pipelineDefinition` 的 定义创建者/定义名称:定义版本

[docker,k8s] 类型的 `task image` 值为容器镜像

#### commands

[os,docker,k8s] 类型的 `task` 声明需要执行的 `shell` 命令

#### type
声明 `task` 的类型，目前分 4 种 [os, docker, k8s, pipeline]

#### actuatorSelector
在 `pipelineDefinition` 中声明全局的 `tag`, 没有声明 `actuatorSelector` 的 `task` 会使用这些全局的 `actuatorSelector`

在 `task` 中声明的 `actuatorSelector`，只能作为当前任务的局部 `actuator`

#### timeout
任务执行的超时时间，超时会自动停止

### 一些概念

#### pipelineDefinition
流水线定义

类似 `docker image`

#### pipeline
流水线定义的 `runtime`

类似 `docker container`

#### actuatorDefinition
声明执行器

目前有 3 类 [os, docker, k8s]。其中有个 tag 字段可以为执行器声明别名

`eventops` 下 `example/actuator` 中有很多 `actuator` 的例子

注意:
1. docker actuator 需要 docker daemon 开启 tcp
2. docker ssh actuator 需要 docker daemon 开启 tcp
3. docker tunnel actuator 需要 docker daemon 开启 tcp
4. docker 使用 ssh docker ip 地址可以使用 127.0.0.1
5. docker 使用 tunnel docker ip 地址可以使用 127.0.0.1
6. k8s actuator 的 config 配置中的 server:http://xxxx:6443 地址需要要能被 eventops 访问。 一般直接拷贝过来的都是域名地址，可能无法解析
7. k8s tunnel actuator 的 config 配置中的 server:http://xxxx:6443 地址需要能被 client 宿主机访问
8. os tunnel actuator 中的 ip 地址和端口要能被 client 宿主机访问

#### actuator
可以看作 task 执行的机器或者容器，或者可以看作 actuatorDefinition

#### triggerDefinition
触发器定义

用于多个流水线和事件进行绑定，可以对事件内容进行过滤，运行流水线的时候还可以将事件的内容传递给流水线

#### task
流水线定义中的任务，执行的最小单位

分 4 种类型 [os, docker, k8s, pipeline] 其中 `pipeline` 类型效果是执行另一个流水线

#### event
事件

用户使用 `eoctl` 发送或者直接调用 `server` 的 `api`
