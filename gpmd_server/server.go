package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zhuangsirui/godist/gpmd"
)

func main() {
	gpmd.Init()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-c
	gpmd.Stop()
	gpmd.Stopped()
}
