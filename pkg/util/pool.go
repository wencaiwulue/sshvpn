package util

import "sync"

var SPool = sync.Pool{
	New: func() any {
		return make([]byte, 20)
	},
}

var MPool = sync.Pool{
	New: func() any {
		return make([]byte, 1024*4)
	},
}
