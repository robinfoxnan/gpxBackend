package common

import (
	"strconv"
	"sync"
	"time"
)

var GlobalSnowFlake *Snowflake = nil

func InitSnow(workerID uint) *Snowflake {
	GlobalSnowFlake = NewSnowflake(workerID)
	return GlobalSnowFlake
}

func NextFileName() string {
	if GlobalSnowFlake == nil {
		return ""
	}

	id := GlobalSnowFlake.NextID()
	filename := strconv.FormatUint(id, 10)
	return filename
}

/*
在这个雪花算法的实现中，生成的唯一ID占用64位，被划分为三个部分：时间戳、工作节点ID和序列号。

时间戳（41位）：
占用位数：从最高位开始的41位。
表示内容：表示生成ID的时间戳，精确到毫秒级别。这个时间戳是相对于算法的开始时间（startTime）的差值。

工作节点ID（10位）：
占用位数：接着时间戳的低位，占用10位。
表示内容：用于标识生成ID的机器（工作节点）的唯一ID。这允许在分布式系统中多个节点同时生成唯一的ID。

序列号（12位）：
占用位数：最后的12位。
表示内容：用于确保在同一毫秒内，同一节点上生成的ID不重复。如果在同一毫秒内生成的ID超过了4096（2^12），则需要等待下一毫秒再生成ID。
总的来说，这个雪花算法的位分配为 41 + 10 + 12 = 63 位，剩下的1位用于符号位（标记正负）。这种分配保证了生成的ID在一定时间内是唯一的，
同时能够支持多个工作节点，每个节点每毫秒可生成4096个不同的ID。
*/
// Snowflake 结构体
type Snowflake struct {
	mu        sync.Mutex
	startTime int64
	workerID  uint
	sequence  uint
	lastTime  int64
}

/*
startTime：表示雪花算法的开始时间，通常是一个固定的时间点，用来减小生成的ID大小。
workerID：标识当前的工作节点（机器）的唯一ID。
sequence：序列号，用来保证在同一毫秒内生成的ID不重复。
lastTime：上一次生成ID的时间戳。
*/

// NewSnowflake 创建一个新的 Snowflake 实例
func NewSnowflake(workerID uint) *Snowflake {
	return &Snowflake{
		startTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano(),
		workerID:  workerID,
		sequence:  0,
		lastTime:  0,
	}
}

/*
获取当前时间戳 currentTime。
检查当前时间戳是否小于上一次生成ID的时间戳 lastTime，如果是，表示时钟回退，抛出异常。
如果当前时间戳等于上一次生成ID的时间戳，则递增序列号 sequence，并检查是否超过最大值（4095），如果超过，等待一段时间，直到下一毫秒再生成ID。
如果当前时间戳大于上一次生成ID的时间戳，则重置序列号为0。
更新 lastTime 为当前时间戳。
使用时间戳、workerID 和序列号生成最终的64位唯一ID。
这样，Snowflake 结构体就可以在分布式环境中生成唯一的ID，其中包含了时间戳、工作节点ID和序列号，以确保在一定时间内和相同工作节点上生成的ID都是唯一的。
*/
// NextID 生成下一个唯一ID
func (sf *Snowflake) NextID() uint64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	currentTime := time.Now().UnixNano()
	if currentTime < sf.lastTime {
		// 如果系统时间倒退，抛出异常或者等待
		//panic("Clock moved backwards. Refusing to generate ID.")
		return 0
	}

	if currentTime == sf.lastTime {
		sf.sequence = (sf.sequence + 1) & 0xfff
		if sf.sequence == 0 {
			// 等待一段时间，确保序列号不重复
			time.Sleep(1 * time.Nanosecond)
			currentTime = time.Now().UnixNano()
		}
	} else {
		sf.sequence = 0
	}

	sf.lastTime = currentTime

	id := uint64((currentTime-sf.startTime)<<22 | int64(sf.workerID)<<12 | int64(sf.sequence))
	return id
}

//func main() {
//	// 创建一个workerID为1的Snowflake实例
//	snowflake := NewSnowflake(1)
//
//	// 生成10个唯一ID并打印
//	for i := 0; i < 10; i++ {
//		id := snowflake.NextID()
//		fmt.Println("Generated ID:", id)
//	}
//}

func generateUniqueID() int64 {
	// This is a simple example; in production, you might want to use a more robust method to generate unique IDs.
	return time.Now().UnixNano() / int64(time.Millisecond)
}
