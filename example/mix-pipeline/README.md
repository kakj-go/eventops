### minio 文件对象准备

1. 下载 minio (地址: `http://www.minio.org.cn/download.shtml#/linux`)
2. 启动 minio (命令: `minio server /root/data`)
3. 下载客户端命令 mc (地址: `http://www.minio.org.cn/download.shtml#/linux`)
4. 使用 mc 命令登录 (命令: `mc alias set kakj_minio http://eventopsIp:eventopsPort accessKey secretKey`)
5. 创建 basePath (命令: `mc mb kakj_minio/eventops`)
6. 创建对象文件并上传 (命令: `echo "hello world input file value" > /root/echo_value && mc cp /root/echo_value kakj_minio/eventops/test`)
7. 注意 eventops 中的 config.yaml 要配置 minio 的信息及 basePath 地址

### 创建 actuator
> 注意下面所有 yaml 文件都要根据实际情况修改, osActuator 对应的宿主机要内置 mc 命令

> 其中以 tunnel 结尾的　yaml 文件需要在宿主机上启动　client 命令，否则可能会导致任务无法创建执行 client

1. eoctl actuator apply -f example/actuator/dockerActuator.yaml
2. eoctl actuator apply -f example/actuator/dockerActuatorSsh.yaml
3. eoctl actuator apply -f example/actuator/dockerActuatorTunnel.yaml
4. eoctl actuator apply -f example/actuator/k8sActuator.yaml
5. eoctl actuator apply -f example/actuator/k8sActuatorTunnel.yaml
6. eoctl actuator apply -f example/actuator/osActuator.yaml
7. eoctl actuator apply -f example/actuator/osActuatorTunnel.yaml

### 创建流水线 
1. eoctl pipeline apply -f pipelineTask2.yaml
2. eoctl pipeline apply -f pipelineTask1.yaml
3. eoctl pipeline apply -f pipelineDefinition.yaml

### 创建触发器
1. eoctl trigger apply -f triggerDefinition.yaml

### 模拟发送事件
1. eoctl event send -f event.yaml


### Dockerfile 文件说明
如果流水线或者触发器有文件传递则需要内置 mc 命令, 该文件是内置 mc 命令的示例
