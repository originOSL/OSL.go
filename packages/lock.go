// name: lock
// description: Sync utilities
// author: Mist
// requires: sync

type OSL_lock struct {
	mu   sync.Mutex
	once map[string]chan struct{}
}

func (l *OSL_lock) Lock(name string) {
	if l == nil {
		return
	}

	for {
		l.mu.Lock()

		if l.once == nil {
			l.once = make(map[string]chan struct{})
		}

		// if no lock exists, acquire it
		if _, ok := l.once[name]; !ok {
			l.once[name] = make(chan struct{})
			l.mu.Unlock()
			return
		}

		// otherwise wait for it to be released
		ch := l.once[name]
		l.mu.Unlock()

		<-ch
	}
}

func (l *OSL_lock) Unlock(name string) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.once == nil {
		return
	}

	if ch, ok := l.once[name]; ok {
		close(ch)
		delete(l.once, name)
	}
}

var lock = &OSL_lock{}
