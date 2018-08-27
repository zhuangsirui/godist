package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/zhuangsirui/godist"
)

func main() {
	//rand.Seed(time.Now().Unix())
	godist.Init(fmt.Sprintf("%d@localhost", rand.Int()))
	godist.Register()
	godist.NewProcess()
	time.Sleep(time.Second * 2)
	godist.Unregister()
}
