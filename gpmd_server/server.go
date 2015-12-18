package main

import (
	"godist/gpmd"
	"os"
	"os/signal"
)

func main() {
	gpmd.Init()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
