package main

import (
	"godist/gpmd"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	gpmd.Init()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-c
	gpmd.Stop()
	gpmd.Stopped()
}
