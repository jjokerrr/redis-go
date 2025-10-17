package main

import (
	"fmt"
	"redis-go/config"
	"redis-go/tcp"
)

const defaultConfigFile = "redis.conf"

func main() {
	config.SetupConfig(defaultConfigFile)
	handler := tcp.MakeHandler()

	_ = tcp.ListenAndServeWithSignal(&tcp.Config{
		Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port)}, handler)

}
