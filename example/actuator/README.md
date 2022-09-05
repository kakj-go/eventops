### 各种 actuator 说明
> 注意所有的 yaml 文件的内容要根据实际情况更改

其中以 tunnel 结尾的　yaml 文件需要在宿主机上启动　client 命令，否则会导致 task 无法创建 actuator client

client 命令中的 id 和 token 配置要设置成 actuator 中 tunnel 的配置

client 命令中的 user 配置要设置成 actuator 的创建人

client 命令中的 connect 配置要设置成 eventops 的地址

> client 命令示例: ./bin/client-linux-amd64 --connect=ws://192.168.0.109:8080/api/dialer/connect --id=runner_tunnel --token=123456 --user=kakj

#### os actuator
os 类型的执行器，如果流水线或者触发器有文件传递，则需要 os 机器上内置 mc 和 curl 命令

#### docker actuator
docker 类型的执行器，如果流水线或者触发器有文件传递，则需要 docker 镜像内置 mc 和 curl 命令

> 注意 docker 需要开启 tcp daemon

#### k8s actuator
k8s 类型的执行器，如果流水线或者触发器有文件传递，则需要 k8s 镜像内置 mc 和 curl 命令

> 注意 kube config 直接拷贝过来的内容，里面的 server:address 可能是 dns 地址，需要改成可访问的地址
