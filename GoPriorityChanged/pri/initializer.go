package pri

import (
	"sync"
)

var (
	Wg sync.WaitGroup
)

func init() {
	initSufMap()
}
