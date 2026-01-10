# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个使用 Go 语言实现的 Redis 简化版本,用于个人学习。该项目实现了 RESP (REdis Serialization Protocol) 协议解析、TCP 服务器和基本的数据库接口。

参考项目: [Redigo](https://github.com/inannan423/redigo)

## 常用命令

### 构建与运行
```bash
# 构建项目
go build -o redis-go

# 运行服务器 (需要 redis.conf 配置文件)
./redis-go

# 或者直接运行
go run main.go
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./resp/parser
go test ./tcp
```

### 开发工具
```bash
# 格式化代码
go fmt ./...

# 检查代码
go vet ./...
```

## 核心架构

### 三层架构设计

1. **TCP 层** (`tcp/`)
   - `server.go`: 实现 TCP 服务器,处理信号、监听端口和连接管理
   - 使用 `ListenAndServeWithSignal()` 启动服务器,支持优雅关闭 (SIGTERM, SIGINT 等)
   - 每个客户端连接在独立的 goroutine 中处理

2. **RESP 协议层** (`resp/`)
   - `parser/parser.go`: 流式解析 RESP 协议,支持单行和多行命令
   - `handler/handler.go`: RESP 请求处理器,管理活跃连接并调用数据库执行命令
   - `connection/connection.go`: 封装客户端连接
   - `reply/`: 各种 RESP 回复类型的实现 (BulkReply, MultiBulkReply, StatusReply 等)

3. **数据库层** (`database/`)
   - `interface/database/database.go`: 定义数据库接口 (Exec, AfterClientClose, Close)
   - `echo_database.go`: 当前实现的简单 Echo 数据库,返回接收到的命令

### 关键数据流

```
Client -> TCP Server -> RESP Parser -> RESP Handler -> Database -> RESP Reply -> Client
```

1. `main.go` 启动 TCP 服务器,使用 `resp/handler` 作为处理器
2. TCP 连接交给 `RespHandler.Handle()` 处理
3. `parser.ParseStream()` 流式解析 RESP 协议命令
4. 解析后的命令通过 `Database.Exec()` 执行
5. 执行结果通过 `Reply.ToBytes()` 序列化后返回客户端

### 配置系统

- `config/config.go`: 使用反射实现的配置文件解析器
- 配置文件格式: `redis.conf` (key-value 格式,空格分隔)
- 全局配置对象: `config.Properties`
- 支持的配置项: bind, port, appendOnly, maxClients, requirePass, databases 等

### 工具库

- `lib/logger/`: 日志系统
- `lib/sync/atomic/`: 原子布尔值实现
- `lib/wait/`: 等待组工具

### Interface 定义

- `interface/tcp/handler.go`: TCP 处理器接口
- `interface/resp/conn.go`: RESP 连接接口
- `interface/resp/reply.go`: RESP 回复接口
- `interface/database/database.go`: 数据库接口

## 开发注意事项

### RESP 协议解析
- RESP 协议支持 5 种类型: 简单字符串 (+), 错误 (-), 整数 (:), 批量字符串 ($), 数组 (*)
- 解析器使用状态机 (`readState`) 处理多行数据
- 所有命令以 `\r\n` 结尾

### 并发安全
- `RespHandler.activeConn` 使用 `sync.Map` 存储活跃连接
- `RespHandler.closing` 使用原子操作控制服务器关闭状态
- 每个客户端连接在独立 goroutine 中处理

### 错误处理
- IO 错误 (如 EOF) 会关闭客户端连接
- 协议解析错误会返回错误回复但保持连接
- 使用 `recover()` 捕获 panic 并记录堆栈

### 扩展数据库实现
要实现新的数据库类型:
1. 在 `database/` 下创建新文件
2. 实现 `interface/database/Database` 接口的三个方法
3. 在 `resp/handler/handler.go` 的 `MakeHandler()` 中替换数据库实例