package license

import (
	"sync"
)

var idsBuffer = &sync.Pool{
	New: func() interface{} {
		return new(int64)
	},
}

type license struct {
	digs int64
	id   int64
	desc *int64
}

func (l license) Close() {
	idsBuffer.Put(l.desc)
	l.desc = nil
	l.digs = -1
}

func newLicense(id, digs int64) license {
	ptr := idsBuffer.Get().(*int64)
	*ptr = digs
	return license{
		digs: digs,
		id:   id,
		desc: ptr,
	}
}
