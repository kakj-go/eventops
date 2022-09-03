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