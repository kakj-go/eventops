# todo 说明
todo 需求优先级分为 `高` `中` `低` 然后还有额外状态 coding, coding 中 [] 代表开发人员

需求完成后就会将内容基于删除线

## coding
1. 高-1 [kakj-go]
2. 高-2 [kakj-go]
3. 高-3 [kakj-go]
4. 高-5 [kakj-go]
5. 高-6 [kakj-go]

## 高
1. 文档补充完全
2. task 的 caches 实现
3. pipeline 定期回收在 actuator 上的资源(容器或者宿主机上的文件)
4. 全面功能性测试
5. 代码结构优化
6. task 上下文传递
7. task 条件跳转执行

## 中
1. eventops server 整个服务拆分，各个拆分的服务考虑横向扩展
2. 单侧覆盖率达到一定的系数
3. 自动化测试覆盖率达到一定的系数
4. 压测数据

## 低
1. 考虑是否和 terraform 一样, 将 pipeline 抽象成一种语言，而放弃用复杂的 yaml 结构
2. eventops 拆分后，pipeline 任务具体执行组件是否支持类似 kubeEdge 的主边模式, 这样能否更好的支持一些极端的银行或者安全性比较高的场景