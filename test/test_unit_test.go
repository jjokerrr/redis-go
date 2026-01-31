package test

import (
	"fmt"
	"redis-go/config"
	"redis-go/database"
	"redis-go/datastruct/dict"
	database2 "redis-go/interface/database"
	"redis-go/resp/connection"
	"testing"
	"time"
)

const defaultConfigFile = "../redis.conf"

// 单测文件

func TestMakeDB(t *testing.T) {
	db := database.NewDB(database.WithIndex(1), database.WithData(dict.MakeSyncDict()))
	fmt.Printf("db:%+v\n", db)
}

func TestPingFunc(t *testing.T) {
	db := database.NewDB(database.WithIndex(1), database.WithData(dict.MakeSyncDict()))
	execReply := db.Exec(nil, [][]byte{[]byte("ping")})
	fmt.Printf("execReply: %s\n", execReply.ToBytes())
}

func TestPutEntityAndGetEntity(t *testing.T) {
	db := database.NewDB(database.WithIndex(1), database.WithData(dict.MakeSyncDict()))
	db.PutEntity("Hello", &database2.DataEntity{Data: "World"})
	entity, exists := db.GetEntity("Hello")
	fmt.Printf("entity: %s, exists: %v\n", entity.Data, exists)
	db.PutIfAbsent("Hello", &database2.DataEntity{Data: "World2"})
	entity, exists = db.GetEntity("Hello")
	fmt.Printf("entity: %s, exists: %v\n", entity.Data, exists)
	db.PutIfExists("Hello", &database2.DataEntity{Data: "World3"})
	entity, exists = db.GetEntity("Hello")
	fmt.Printf("entity: %s, exists: %v\n", entity.Data, exists)
	db.Remove("Hello")
	entity, exists = db.GetEntity("Hello")
	fmt.Printf("entity: %s, exists: %v\n", entity.Data, exists)
	db.PutEntity("Hello", &database2.DataEntity{Data: "World4"})
	db.Flush()
	entity, exists = db.GetEntity("Hello")
	fmt.Printf("entity: %s, exists: %v\n", entity.Data, exists)
}

func TestAof(t *testing.T) {

	config.SetupConfig(defaultConfigFile)
	standaloneDatabase := database.NewStandaloneDatabase()
	fakeConn := &connection.Connection{}
	standaloneDatabase.Exec(fakeConn, [][]byte{[]byte("set"), []byte("hello"), []byte("world")})
	reply := standaloneDatabase.Exec(fakeConn, [][]byte{[]byte("get"), []byte("hello")})
	fmt.Printf("reply: %s\n", reply.ToBytes())
	time.Sleep(time.Second * 5)
}
