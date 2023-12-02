package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const nodeIDEnv = "SNOWFLAKE_NODE_ID"

//go:generate gogen option -n Node -s node,nodeBits,globalFlag,storage,epoch --with-error --with-init -p _
type Node struct {
	mu sync.Mutex

	storage    Storage
	epoch      time.Time
	time       int64
	globalFlag bool
	node       uint64
	step       uint64

	nodeBits uint8
	stepBits uint8
	nodeMax  uint64
	stepMax  uint64

	nodeShift       uint8
	globalFlagShift uint8
	timeShift       uint8
}

type Storage interface {
	Get() (t int64, err error)
	Save(t int64) error
}

func (n *Node) init() error {
	if n.epoch.IsZero() {
		n.epoch = time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local)
	}
	if n.time == 0 {
		n.time = time.Since(n.epoch).Milliseconds()
	}
	if n.nodeBits == 0 {
		n.nodeBits = nodeBits
	}
	// 文件中读取时间
	if n.storage != nil {
		t, err := n.storage.Get()
		if err != nil {
			return err
		}
		if t > n.time {
			n.time = t
		}
	}

	// 第一位保留
	var bits uint8 = 64 - 1
	n.stepBits = bits - timeBits - globalFlagBits - n.nodeBits
	n.nodeMax = calcMax(n.nodeBits)
	n.stepMax = calcMax(n.stepBits)
	n.nodeShift = n.stepBits
	n.globalFlagShift = n.stepBits + n.nodeBits
	n.timeShift = n.stepBits + n.nodeBits + globalFlagBits

	if n.node == 0 {
		nodeStr := os.Getenv(nodeIDEnv)
		if nodeStr != "" {
			node, err := strconv.Atoi(nodeStr)
			if err != nil {
				return err
			}
			n.node = uint64(node)
		}
	}

	if n.node < 0 || n.node > n.nodeMax {
		panic(fmt.Sprintf("node number must be between 0 and %d, but this node is %d", n.nodeMax, n.node))
	}
	return nil
}

func (n *Node) NodeBits() uint8 {
	return n.nodeBits
}

func (n *Node) StepBits() uint8 {
	return n.stepBits
}

func (n *Node) Gen() (ID, error) {
	return n.gen(1)
}

func (n *Node) MustGen() ID {
	id, err := n.Gen()
	if err != nil {
		panic(err)
	}
	return id
}

func (n *Node) Alloc(step uint64) (ID, error) {
	return n.gen(step)
}

func (n *Node) MustAlloc(step uint64) ID {
	id, err := n.Alloc(step)
	if err != nil {
		panic(err)
	}
	return id
}

func (n *Node) gen(step uint64) (ID, error) {
	if step > n.stepMax {
		return 0, fmt.Errorf("step(%d) can't greater than calcMax(%d)", step, n.stepMax)
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.epoch).Milliseconds()
begin:
	// 发生了时钟回拨或处在同一毫秒
	if now <= n.time {
		// 当前毫秒数足够分配step
		if n.step+step <= n.stepMax {
			n.step += step
		} else {
			// 当前毫秒数无法分配足够的step，借用下一毫秒的。
			n.step = 0
			n.time += 1
			goto begin
		}
	} else {
		// 当前请求时间已大于n.time，则从当前时间开始分配step
		n.step = 0
		n.time = now
		if n.storage != nil {
			if err := n.storage.Save(now); err != nil {
				return 0, err
			}
		}
		goto begin
	}
	var globalFlag uint64
	if n.globalFlag {
		globalFlag = 1
	}
	return ID((uint64(n.time) << n.timeShift) | (globalFlag << n.globalFlagShift) | (n.node << n.nodeShift) | n.step), nil
}

func calcMax(bits uint8) uint64 {
	i := -1 ^ (-1 << bits)
	return uint64(i)
}

// DefaultNode 默认阶段
var DefaultNode = MustNew()

func Gen() (ID, error) {
	return DefaultNode.Gen()
}

func MustGen() ID {
	return DefaultNode.MustGen()
}

func Alloc(step uint64) (ID, error) {
	return DefaultNode.Alloc(step)
}

func MustAlloc(step uint64) ID {
	return DefaultNode.MustAlloc(step)
}
