// name: lock
// description: Sync utilities
// author: Mist
// requires: sync

type Sync struct{}

func (Sync) Mutex() *sync.Mutex {
	return &sync.Mutex{}
}

func (Sync) RWMutex() *sync.RWMutex {
	return &sync.RWMutex{}
}

var lock = Sync{}