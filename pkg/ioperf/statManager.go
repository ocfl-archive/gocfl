package ioperf

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"sync/atomic"
)

type StatManager struct {
	start      timestamp.Timestamp
	readBytes  uint64
	writeBytes uint64
}

func (sm *StatManager) AddRead(size uint64) {
	atomic.AddUint64(&sm.readBytes, size)
}

func (sm *StatManager) AddWrite(size uint64) {
	atomic.AddUint64(&sm.writeBytes, size)
}
