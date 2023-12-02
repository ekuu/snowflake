package snowflake

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	timeBits       uint8 = 41
	globalFlagBits uint8 = 1
	nodeBits       uint8 = 6
)

// ID
// 从高到低依次为:
// 1位保留位
// 41位为毫秒级时间戳
// 1位标识是否是作为服务提供全局唯一的ID：1为是，0为否(本地生成)
// nodeBits位标识机器
// 其余位标识每毫秒计数

type ID uint64

func (i ID) Uint64() uint64 {
	return uint64(i)
}

func (i ID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i ID) Hex() string {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return hex.EncodeToString(bytesBuffer.Bytes())
}

func (i ID) Base32() string {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(bytesBuffer.Bytes())

}

func (i ID) Base32Lower() string {
	return strings.ToLower(i.Base32())
}

func (i ID) UnixMilli(n *Node) int64 {
	return int64((i.Uint64() >> n.timeShift) & calcMax(timeBits))
}

func (i ID) Time(n *Node) time.Time {
	return n.epoch.Add(time.Millisecond * time.Duration(i.UnixMilli(n)))
}

func (i ID) IsGlobal(n *Node) bool {
	return (i.Uint64()>>n.globalFlagShift)&1 == 1
}

func (i ID) Node(n *Node) uint64 {
	return (i.Uint64() >> n.nodeShift) & calcMax(n.nodeBits)
}

func (i ID) Step(n *Node) uint64 {
	return i.Uint64() & calcMax(n.stepBits)
}

func byteOrder() binary.ByteOrder {
	// 占4byte 转换成16进制 0x00 00 00 01
	var value int32 = 1
	// 大端(16进制)：00 00 00 01
	// 小端(16进制)：01 00 00 00
	pointer := unsafe.Pointer(&value)
	pb := (*byte)(pointer)
	if *pb == 1 {
		return binary.LittleEndian
	} else {
		return binary.BigEndian
	}
}
