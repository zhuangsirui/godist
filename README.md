# Godist

Go Distributed System 是一个允许在一个分布式 Go 集群中，方便地在任何一个 `Goroutine` 中找到在集群中任意节点的 `Goroutine` ，并于之相互通信的分布式系统基础组件。

Godist 参照 Erlang 的思路，由两部分组成： `GPMD` 和 `Agent` 。

第一部分 `GPMD` 是一个守护进程，每台物理服务器需要启动一个该守护进程，用于监督管理本服务器的 Go 进程中的 `Agent` 。

第二部分是在每个 Go 进程中初始化的 `Agent` 。该进程中的 `Routine` 如需要在集群中与其他其他任意 `Routine` 交互，则需要在本进程中的 `Agent` 注册一条消息。

## GPMD

Go Process Mapping Daemon

## Agent

Goroutine agent for distributed system
