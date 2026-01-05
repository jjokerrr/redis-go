package database

import "strings"

// 命令集合，保存所有命令的元信息
var cmdTable = make(map[string]*command)

// command 命令元信息，包含命令的名称，命令需要的参数个数，命令执行函数
type command struct {
	name  string
	exec  ExecFunc
	arity int
}

// RegisterCommand 命令注册方法
func RegisterCommand(name string, exec ExecFunc, arity int) {
	name = strings.TrimSpace(strings.ToLower(name)) // 做一下兼容性处理
	cmdTable[name] = &command{
		exec:  exec,
		arity: arity,
	}
}
