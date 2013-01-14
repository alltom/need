package need

import "sync"

// caches all things in memory, forever
type MemoryProvider struct {
	cache map[string]interface{}
	mutex sync.Mutex
}

func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{make(map[string]interface{}), sync.Mutex{}}
}

func (p *MemoryProvider) ValueRequested(key string) (interface{}, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if value, ok := p.cache[key]; ok {
		return value, true
	}
	return nil, false
}

func (p *MemoryProvider) ValueArrived(key string, value interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache[key] = value
}
