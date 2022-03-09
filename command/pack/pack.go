package pack

import (
	"YaeDisk/logx"
	"fmt"
	"github.com/vmihailenco/msgpack"
)

func EncodePack(arg interface{}) []byte {
	b, err := msgpack.Marshal(arg) // 将结构体转化为二进制流
	if err != nil {
		logx.Warn(fmt.Sprintf("msgpack marshal failed,err:%v", err))
		return nil
	}
	return b
}

func XDecodePack(arg []byte) interface{} {
	var p2 interface{}
	err := msgpack.Unmarshal(arg, &p2) // 将二进制流转化回结构体
	if err != nil {
		logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
		return nil
	}
	return p2
}
