package test

import (
	"fmt"
	"redis-go/database"
	"redis-go/datastruct/dict"
	"testing"
)

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
