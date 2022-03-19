package utils

import (
	"encoding/binary"
	"math"
	"strconv"
)

func UInt64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func BytesToInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func UInt64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func StringToUInt64(s string) uint64 {
	if parseUint, err := strconv.ParseUint(s, 10, 64); err != nil {
		return math.MaxUint64
	} else {
		return parseUint
	}
}
