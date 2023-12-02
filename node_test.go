package snowflake

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNode_Gen(t *testing.T) {
	var id ID
	//n := MustNew()
	n := MustNew(WithGlobalFlag(true), WithNode(1))
	fmt.Println(len("11101000100100110101000011000101101000001000011111010000"))
	id = n.MustGen()
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
	id = n.MustGen()
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
	id = n.MustGen()
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
	id = n.MustGen()
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())

	// 32767
	id = n.MustAlloc(32767)
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
	id = n.MustAlloc(1000)
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
	id = n.MustAlloc(1000)
	t.Logf("%d,%s,%d,%b,%s,%s\n", id, id.Time(n), id.Node(n), id, id.Hex(), id.Base32Lower())
}

func TestNode_decode(t *testing.T) {
	b, err := base32.HexEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper("110pe9oak2004"))
	if err != nil {
		t.Fatal(err)
	}

	id := ID(binary.BigEndian.Uint64(b))
	l, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id.Node(DefaultNode), id.Time(DefaultNode).In(l))
}
