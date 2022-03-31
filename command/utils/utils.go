package utils

import (
	"encoding/binary"
	"strconv"
)

// UInt64ToBytes
// 将 uint64 转换为 []byte
func UInt64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

// BytesToInt64
// 将 []byte 转换为 uint64
func BytesToInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

// UInt64ToString
// 将 uint64 转换为 string
func UInt64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// StringToUInt64
// 将 string 转换为 uint64
func StringToUInt64(s string) (bool, uint64) {
	if parseUint, err := strconv.ParseUint(s, 10, 64); err != nil {
		return false, 0
	} else {
		return true, parseUint
	}
}
