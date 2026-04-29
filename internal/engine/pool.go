package engine

import "sync"

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 1600)
	},
}
