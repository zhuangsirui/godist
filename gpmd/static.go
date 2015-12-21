package gpmd

var (
	_port    uint16 = 2613
	_host           = ""
	_manager *Manager
)

// 设置 GPMD 的监听端口。默认 2613 。
func SetPort(port uint16) {
	_port = port
}

// 设置 GPMD 的监听地址。默认为空。
func SetHost(host string) {
	_host = host
}

// 初始化 GPMD 服务。
func Init() {
	_manager = New(_host, _port)
	_manager.Serve()
}

func Stop() {
	_manager.Stop()
}
