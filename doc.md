## 一些概念

### pipelineDefinition
流水线定义

类似 `docker image`

### pipeline
流水线定义的 `runtime` 

类似 `docker container`

### actuatorDefinition
执行器声明

目前有 3 类 [os, docker, k8s]。其中有个 tag 字段可以为执行器声明别名

### actuator
可以看作 task 执行的机器或者容器，或者可以看作 actuatorDefinition

### triggerDefinition
触发器定义

用于多个流水线和事件进行绑定，可以对事件内容进行过滤，运行流水线的时候还可以将事件的内容传递给流水线

### task
流水线定义中的任务，执行的最小单位

分 4 种类型 [os, docker, k8s, pipeline] 其中 `pipeline` 类型效果是执行另一个流水线

### event
事件

用户使用 `eoctl` 发送或者直接调用 `server` 的 `api`

# 文档

## pipelineDefinition
yaml 字段描述

> 注意: 如果流水线入参全局变量和出参，或者任务的出参如果存在文件类型的值引用，则 server 的 config.yaml 中需要配置 minio, 然后运行任务的宿主机或者容器需要内置 mc(minio client) 命令
> 

### name
声明定义的名称

### version
声明定义的版本

### actuatorSelector
声明全局的 `tag`, 没有声明 `actuatorSelector` 的 `task` 会使用这些全局的 `actuatorSelector`

### inputs
`inputs` 声明该流水线运行时需要那些入参

`task` 可以使用 `${{ inputs.inputName }}` 来使用入参的值

### outputs
`outputs` 声明该定义会产生那些出参及值是那些任务的出参

### contexts
定义中声明全局变量，`task` 的 `outputs` 字段 `setToContexts` 进行设值

`task` 可以使用 `${{ contexts.contextName }}` 来使用值

### dag
声明任务的依赖关系和运行流程, 关系应该是有向无环图

### tasks
任务列表

#### image

[pipeline] 类型的 `task` 的 `image` 值应该是 `pipelineDefinition` 的定义

例子: `pipelineDefinitionCreater/pipelineDefinitionName:pipelineDefinitionVersion`

[docker, k8s] 类型的 `task` 的 `image` 值为容器镜像

[os] 类型 `task` 没有 `image`

#### commands

[os,docker, k8s] 类型的 `task` 声明需要执行的 `shell` 命令

[pipeline] 类型的 `task` 没有该字段

#### type
声明 `task` 的类型，目前分 4 种 [os, docker, k8s, pipeline]

#### actuatorSelector
在 `task` 中声明的 `actuatorSelector`，只能作为当前任务的局部 `actuator`

#### timeout
任务执行的超时时间，超时会自动停止

#### inputs

[pipeline] 类型的 `task` 中的 `inputs` 代表运行定义传递入参的值

[os, docker, k8s] 类型的 `task` 没有该值

#### outputs

[os, docker, k8s] 类型的 `task` 出参可以声明使用那些环境变量的值或者那个绝对路径的文件

[pipeline] 类型的 `task` 只能引用流水线所具有的出参

`outputs` 字段 `setToContexts` 对 `contexts` 进行设值

`task` 可以使用 `${{ outputs.taskName.taskOutputName }}` 来使用值

```yaml
version: 1.0 # 声明流水线的版本
name: mix-pipeline # 声明流水线的名称

actuatorSelector: # 声明使用那些 tag 的 actuator
  tags:
    - os-runner
    - docker-runner
    - k8s-runner

inputs: # 声明流水线入参
  - name: echo_env_value # 入参的名称
    type: env # 值类型
  - name: echo_file_value # 入参的名称
    type: file # 文件类型

contexts: # 声明全局变量
  - name: context_env_value # 全局变量的名称
    type: env # 值类型
  - name: context_file_value
    type: file # 文件类型

# 任务的 dag 依赖描述, 所有 tasks 都需要在 dag 中描述其运行依赖关系
dag:
  - name: os-output-context
    needs: # 该任务需要等待 needs 的任务执行完成才进行
      - root  # needs: - root 是必须存在的, 它是流水线执行的起点
  - name: pipeline-echo-context
    needs:
      - os-output-context

# 任务列表
tasks:
  # [os, docker, k8s] 任务出参可以引用环境变量和文件，文件可以使用绝对地址或者环境变量最终指向的绝对地址
  # 出参可以设置到全局变量中
  # commands 中可以使用 ${{ inputs.inputName }} ${{ contexts.contextName }} ${{ outputs.taskName.outputName }} 等表达式来引用值
  # os 类型的 image 可以为空
  - alias: os-output-context # 任务别名
    type: os # 任务类型
    outputs: # 任务的出参
      - name: context_env_output # 任务出参名称
        value: contextEnvValue # 任务出参引用那个环境变量的值
        type: env # 值类型
        setToContext: context_env_value # 出参设置到全局变量 context_env_value 中
      - name: context_file_output
        value: $contextFile 或者 /root/contextFile # 使用环境变量或者使用绝对路径，环境变量值最终也应该是绝对路径
        type: file # 文件类型
        setToContext: context_file_value # 出参设置到全局变量 context_file_value 中
    commands:
      - echo "context file value" > contextFile # 使用 shell 创建文件
      - export contextEnvValue='context env value' # 使用 shell 创建环境变量 contextEnvValue
      - export contextFile=`pwd`/contextFile # 使用 shell 创建环境变量 contextFile
      
      - echo "${{ inputs.echo_env_value }}" # 使用流水线定义的入参
      - cat "${{ inputs.echo_file_value }}" # 使用流水线定义的文件入参
      - echo "${{ outputs.xxx.env }}" # 使用某个任务的出参
      - cat "${{ outputs.xxx.file }}" # 使用某个任务的文件出参
      - echo "${{ contexts.context_env_value }}" # 使用全局变量的值
      - cat "${{ contexts.context_env_value }}" # 使用全局变量的文件
  
  # [k8s, docker] 类型的任务 image 字段值应该是容器镜像
  - alias: k8s-runner
    type: k8s 或者 docker
    image: kakj/mc # 容器的镜像
    commands:
      - echo "k8s and docker"

 # [pipeline] 类型任务 inputs 的 value 可以使用 ${{ inputs.inputName }} ${{ contexts.contextName }} ${{ outputs.taskName.outputName }} 等表达式来引用值
 # [pipeline] 类型任务的 outputs 是引用流水线定义的出参
 # [pipeline] 类型任务的 image 结构:  流水线定义创建者/流水线定义名称:流水线定义版本
  - alias: pipeline-echo-input 
    type: pipeline 
    image: kakj/mix-pipeline-task1:1.0 # [pipeline] 类型的任务对应流水线定义的描述
    outputs: # 出参
      - name: pipeline-env-output # 出参的名称
        value: docker-env-output # 出参引用该任务流水线定义中的出参的名称
        type: env # 值类型
        setToContext: context_env_value
      - name: pipeline-file-output # 出参的名称
        value: docker-file-output # 出参引用该任务流水线定义中的出参的名称
        type: file # 文件类型
        setToContext: context_file_value
    inputs: # 传入给流水线定义的入参
      - name: env1 # 流水线定义的入参名称
        value: ${{ inputs.echo_env_value }} # 使用当前流水线定义的入参的值
      - name: file1 
        value: ${{ inputs.echo_file_value }} # 文件类型的值
      - name: env2 
        value: ${{ contexts.context_env_value }} # 当使用前流水线定义的全局变量的值
      - name: file2 
        value: ${{ contexts.context_file_value }} # 文件类型的值
      - name: env3 
        value: ${{ outputs.os-output-context.context_env_output }} # 使用任务出参的值
      - name: file3 
        value: ${{ outputs.os-output-context.context_file_output }} # 文件类型的值       

outputs:
  - name: outputA # 流水线出参的名称
    value: ${{ outputs.pipeline-echo-output.pipeline-env-output }} # 流水线出参引用那个任务的出参
  - name: outputB # 流水线出参的名称
    value: ${{ outputs.os-output-context.context_env_output }} # 流水线出参引用那个任务的出参
```

## actuatorDefinition
流水线执行会根据 `tag` 获取定义, 然后根据定义创建 `client`, 任务使用 `client` 执行命令

> 注意
> 1. docker 需要 daemon 开启 tcp 连接，如果不是一台机器还需要开放端口或者防火墙
> 2. kubernetes config 文件直接拷贝过来里面的地址可能要改下，如果是 dns 的 server 地址可能是无法解析的
> 3. os kubernetes 和 docker 需要 3 选 1 进行配置
> 4. 只要开启了 tunnel 配置, 各个连接的地址只要是 client 工具运行的主机能访问的地址即可
> 5. `https://github.com/kakj-go/eventops/tree/master/example/actuator` 下有各个连接的例子
> 6. tag 字段是必须的, 名称不会被流水线定义使用

```yaml
name: docker_runner_definition # 定义的名称

docker:
  ip: 127.0.0.1 # docker server ip 地址
  port: 2375 # docker server 端口
  ssh: # docker 配置了 ssh 代表开启 ssh 通道连接
    user: root # ssh 的用户名
    ip: 127.0.0.1 # ssh 机器的 ip
    password: 123456 # ssh 机器的密码

os:
  user: root # ssh 的用户名
  ip: 127.0.0.1 # ssh 机器的 ip
  password: 123456 # ssh 机器的密码
  
kubernetes:
  config: "kube config" # k8s 的 kube config 文件

tunnel:
  clientId: docker_runner # client 命令 --id=xxx 启动中声明的值
  clientToken: 123456 # client 命令 --token=xxx 启动中声明的值  
  
tags:
  - docker_runner_tag # 定义的别名
```

## event
> 注意: 如果事件中 files 中存在值，则需要 server 的 config.yaml 配置 minio, 然后运行任务的宿主机或者容器需要内置 mc(minio client) 命令

事件是一个固定格式的 json, 下面是该 json 的 yaml 结构

```yaml
name: eoctl-run # 事件名称
version: 1.0 # 事件版本

values: # 值类型 map[string]string 格式
  user: kakj
  email: 2357431193@qq.com

files: # 文件类型 map[string]object 格式
  testFile: # 文件的 key 
    value: test/echo_value # 对象的地址会拼接 server config.yaml 中声明的 basePath 的地址
    type: minio # 暂时只支持 minio 类型的值类型

labels: # 标签, 最好用作过滤用 map[string]string 结构
  user: kakj
  email: 2357431193@qq.com

timestamp: 1661422308 # 事件产生的时间
users: ["kakj"] # 事件只同意那些用户的触发器使用
```

## triggerDefinition
> 注意: 如果流水线入参存在文件类型的值引用，则 server 的 config.yaml 中需要配置 minio, 然后运行任务的宿主机或者容器需要内置 mc(minio client) 命令

触发器和事件匹配需要 4 个条件

1. 触发器中声明的事件名称创建者和版本都需要和事件匹配上
2. 事件的 users 中需要有该触发器的创建人
3. 全局 filters 取值表达式取到的值和 matches 中的值需要有一个匹配上
4. 流水线 filters 取值表达式取到的值和 matches 中的值需要有一个匹配上

```
取值表达式使用 `github.com/tidwall/gjson` 库

列子: 

event 的 json 结构:
{
  values: {
    name: 123
  }
}
表达式: values.name = 123
```

```yaml
name: trigger-definition-name # 触发器定义名称

# 注意事件和触发器是根据下面这三个字段匹配才会触发
eventName: eoctl-run # 事件的名称
eventCreater: kakj # 事件的创建人
eventVersion: 1.0 # 事件的版本

# 触发器要触发的流水线
pipelines:
  - image: kakj/hello-world:1.0 # 结构: 流水线定义创建人/流水线定义名称:流水线定义版本
    filters: # 该流水线的事件过滤
      - expr: values.name # 使用 json 取值表达式从 event 的 json 中取值
        matches: # 
          - kakj # 是否匹配 
          - kakj-go # 是否匹配
    inputs: # 传递给流水线的入参
      - name: input_name # 传递给流水线的入参名称
        value: values.name # 传递的值，使用 json 取值表达式从 event 的 json 中取值
      - name: input_file
        value: files.testFile

filters: # 全局事件过滤器
  - expr: values.email # json 取值表达式从 event 的 json 中取值
    matches:
      - kakj # 是否匹配
```
