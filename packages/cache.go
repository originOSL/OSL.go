// name: cache
// description: In-memory LRU cache utilities
// author: roturbot
// requires: time, sync

type Cache struct {
	lru      *lru
	ttl      time.Duration
	metadata map[string]int64
}

type lruItem struct {
	key   string
	value any
	prev  *lruItem
	next  *lruItem
}

type lru struct {
	capacity int
	items    map[string]*lruItem
	head     *lruItem
	tail     *lruItem
	mu       sync.RWMutex
}

func newLRU(capacity int) *lru {
	return &lru{
		capacity: capacity,
		items:    make(map[string]*lruItem),
		head:     nil,
		tail:     nil,
	}
}

func (l *lru) get(key string) any {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		l.moveToFront(item)
		return item.value
	}

	return nil
}

func (l *lru) set(key string, value any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		item.value = value
		l.moveToFront(item)
		return
	}

	item := &lruItem{key: key, value: value}
	l.items[key] = item
	l.addToFront(item)

	if len(l.items) > l.capacity {
		l.removeTail()
	}
}

func (l *lru) remove(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		l.removeItem(item)
		delete(l.items, key)
	}
}

func (l *lru) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.items = make(map[string]*lruItem)
	l.head = nil
	l.tail = nil
}

func (l *lru) size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}

func (l *lru) keys() []any {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]any, 0, len(l.items))
	for key := range l.items {
		keys = append(keys, key)
	}
	return keys
}

func (l *lru) has(key string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, exists := l.items[key]
	return exists
}

func (l *lru) moveToFront(item *lruItem) {
	l.removeItem(item)
	l.addToFront(item)
}

func (l *lru) addToFront(item *lruItem) {
	item.prev = nil
	item.next = l.head

	if l.head != nil {
		l.head.prev = item
	}
	l.head = item

	if l.tail == nil {
		l.tail = item
	}
}

func (l *lru) removeItem(item *lruItem) {
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		l.head = item.next
	}

	if item.next != nil {
		item.next.prev = item.prev
	} else {
		l.tail = item.prev
	}
}

func (l *lru) removeTail() {
	if l.tail != nil {
		delete(l.items, l.tail.key)
		l.removeItem(l.tail)
	}
}

func (Cache) create(capacity any, ttl any) *Cache {
	cap := int(OSLcastNumber(capacity))
	if cap <= 0 {
		cap = 100
	}

	ttlDuration := time.Duration(OSLcastNumber(ttl)) * time.Second
	if ttlDuration <= 0 {
		ttlDuration = 5 * time.Minute
	}

	return &Cache{
		lru:      newLRU(cap),
		ttl:      ttlDuration,
		metadata: make(map[string]int64),
	}
}

func (Cache) createDefault() *Cache {
	return Cache.create(100, 300)
}

func (c *Cache) set(key any, value any) bool {
	if c == nil || c.lru == nil {
		return false
	}

	keyStr := OSLtoString(key)
	c.lru.set(keyStr, value)
	c.metadata[keyStr] = time.Now().Unix()

	return true
}

func (c *Cache) get(key any) any {
	if c == nil || c.lru == nil {
		return nil
	}

	keyStr := OSLtoString(key)

	expiry, exists := c.metadata[keyStr]
	if exists && time.Now().Unix()-expiry > int64(c.ttl.Seconds()) {
		c.delete(keyStr)
		return nil
	}

	return c.lru.get(keyStr)
}

func (c *Cache) getOrSet(key any, value any) any {
	if c.has(key) {
		return c.get(key)
	}
	c.set(key, value)
	return value
}

func (c *Cache) getOrSetFunc(key any, fn any) any {
	if c.has(key) {
		return c.get(key)
	}

	result := OSLcallFunc(fn, nil, []any{})
	c.set(key, result)
	return result
}

func (c *Cache) delete(key any) bool {
	if c == nil || c.lru == nil {
		return false
	}

	keyStr := OSLtoString(key)
	delete(c.metadata, keyStr)
	c.lru.remove(keyStr)

	return true
}

func (c *Cache) clear() bool {
	if c == nil || c.lru == nil {
		return false
	}

	c.lru.clear()
	c.metadata = make(map[string]int64)

	return true
}

func (c *Cache) has(key any) bool {
	if c == nil || c.lru == nil {
		return false
	}

	keyStr := OSLtoString(key)

	expiry, exists := c.metadata[keyStr]
	if !exists || time.Now().Unix()-expiry > int64(c.ttl.Seconds()) {
		c.delete(keyStr)
		return false
	}

	return c.lru.has(keyStr)
}

func (c *Cache) size() int {
	if c == nil || c.lru == nil {
		return 0
	}

	c.cleanupExpired()
	return c.lru.size()
}

func (c *Cache) keys() []any {
	if c == nil || c.lru == nil {
		return []any{}
	}

	c.cleanupExpired()
	return c.lru.keys()
}

func (c *Cache) values() []any {
	if c == nil || c.lru == nil {
		return []any{}
	}

	c.cleanupExpired()

	keys := c.lru.keys()
	values := make([]any, len(keys))

	for i, key := range keys {
		values[i] = c.get(key)
	}

	return values
}

func (c *Cache) entries() map[string]any {
	if c == nil || c.lru == nil {
		return map[string]any{}
	}

	c.cleanupExpired()

	result := make(map[string]any)
	keys := c.lru.keys()

	for _, key := range keys {
		result[OSLtoString(key)] = c.get(key)
	}

	return result
}

func (c *Cache) cleanupExpired() {
	if c == nil || c.lru == nil {
		return
	}

	now := time.Now().Unix()

	for key, expiry := range c.metadata {
		if now-expiry > int64(c.ttl.Seconds()) {
			c.delete(key)
		}
	}
}

func (c *Cache) setTTL(key any, ttl any) bool {
	if c == nil {
		return false
	}

	keyStr := OSLtoString(key)
	if !c.lru.has(keyStr) {
		return false
	}

	ttlDuration := time.Duration(OSLcastNumber(ttl)) * time.Second
	c.metadata[keyStr] = time.Now().Unix() + int64(ttlDuration.Seconds()) - int64(c.ttl.Seconds())

	return true
}

func (c *Cache) getTTL(key any) int {
	if c == nil || c.lru == nil {
		return -1
	}

	keyStr := OSLtoString(key)
	expiry, exists := c.metadata[keyStr]
	if !exists {
		return -1
	}

	remaining := int(float64(expiry + int64(c.ttl.Seconds()) - time.Now().Unix()))

	if remaining < 0 {
		c.delete(keyStr)
		return 0
	}

	return remaining
}

func (c *Cache) stats() map[string]any {
	if c == nil || c.lru == nil {
		return map[string]any{
			"size":     0,
			"capacity": 0,
			"ttl":      0,
			"hits":     0,
			"misses":   0,
		}
	}

	c.cleanupExpired()

	return map[string]any{
		"size":     c.size(),
		"capacity": c.lru.capacity,
		"ttl":      int(c.ttl.Seconds()),
		"keys":     c.keys(),
	}
}

func (c *Cache) setMany(data map[string]any) bool {
	if c == nil {
		return false
	}

	for k, v := range data {
		c.set(k, v)
	}

	return true
}

func (c *Cache) getMany(keys []any) map[string]any {
	if c == nil {
		return map[string]any{}
	}

	result := make(map[string]any)
	for _, key := range keys {
		if value := c.get(key); value != nil {
			result[OSLtoString(key)] = value
		}
	}

	return result
}

func (c *Cache) deleteMany(keys []any) bool {
	if c == nil {
		return false
	}

	for _, key := range keys {
		c.delete(key)
	}

	return true
}

func (c *Cache) filter(fn any) map[string]any {
	if c == nil || c.lru == nil {
		return map[string]any{}
	}

	c.cleanupExpired()

	entries := c.entries()
	result := make(map[string]any)

	for key, value := range entries {
		shouldInclude := OSLcastBool(OSLcallFunc(fn, nil, []any{key, value}))
		if shouldInclude {
			result[key] = value
		}
	}

	return result
}

func (c *Cache) map(fn any) map[string]any {
	if c == nil || c.lru == nil {
		return map[string]any{}
	}

	c.cleanupExpired()

	entries := c.entries()
	result := make(map[string]any)

	for key, value := range entries {
		mappedValue := OSLcallFunc(fn, nil, []any{key, value})
		result[key] = mappedValue
	}

	return result
}

func (c *Cache) reduce(initial any, fn any) any {
	if c == nil || c.lru == nil {
		return initial
	}

	c.cleanupExpired()

	result := initial
	entries := c.entries()

	for _, value := range entries {
		result = OSLcallFunc(fn, nil, []any{result, value})
	}

	return result
}

func (c *Cache) foreach(fn any) bool {
	if c == nil || c.lru == nil {
		return false
	}

	c.cleanupExpired()

	entries := c.entries()
	for key, value := range entries {
		OSLcallFunc(fn, nil, []any{key, value})
	}

	return true
}

type CacheEntry struct {
	Key   string
	Value any
	TTL   int
}

func (c *Cache) toArray() []CacheEntry {
	if c == nil || c.lru == nil {
		return []CacheEntry{}
	}

	c.cleanupExpired()

	entries := c.entries()
	result := make([]CacheEntry, 0, len(entries))

	for key, value := range entries {
		ttl := c.getTTL(key)
		result = append(result, CacheEntry{
			Key:   key,
			Value: value,
			TTL:   ttl,
		})
	}

	return result
}

var cache = Cache{}
