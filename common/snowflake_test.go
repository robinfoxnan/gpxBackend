package common

import (
	"testing"
)

func TestSnow(t *testing.T) {
	InitSnow(1)
	idMap := make(map[uint64]string)
	for i := 0; i < 10000; i++ {
		id := GlobalSnowFlake.NextID()
		_, exists := idMap[id]
		if exists {
			//fmt.Println("Generated same ID:", id)
			t.Errorf("GlobalSnowFlake.NextID returned sameid %d", id)
		} else {
			idMap[id] = " "
		}
	}
}
