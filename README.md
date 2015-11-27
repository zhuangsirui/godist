# Godist

Go Distributed System 是一个允许在一个分布式 Go 集群中，方便地在任何一个 `Goroutine` 中找到在集群中任意节点的 `Goroutine` ，并于之相互通信的分布式系统基础组件。

Godist 参照 Erlang 的思路，由两部分组成： `GPMD` 和 `Agent` 。

第一部分 `GPMD` 是一个守护进程，每台物理服务器需要启动一个该守护进程，用于监督管理本服务器的 Go 进程中的 `Agent` 。

第二部分是在每个 Go 进程中初始化的 `Agent` 。该进程中的 `Routine` 如需要在集群中与其他其他任意 `Routine` 交互，则需要在本进程中的 `Agent` 注册一条消息。

## GPMD

Go Process Mapping Daemon

GPMD 启动以后监听一个默认端口 2613 。当一个本地节点被启动的时候， Agent 会向 GPMD 发送注册请求， GPMD 则会在收到请求之后，将该节点的端口以及其他信息保存，以便当其他节点询问时，将该节点的端口以及其他信息告知对方。

## Agent

Goroutine agent for distributed system

Agent 启动之后会先监听一个随机端口，接着会向本地的 GPMD 注册自身信息。随后，当向当节点指定在分布式系统中的第一个其他节点时，会向目标节点的 GPMD 进行询问，找到目标节点的随机端口之后，向目标节点发起连接请求。请求通过之后，会建立一个 TCP 长连接用于之后 Goroutine 之间的消息发送。
