package need_redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

// caches with redis
// only works with strings
type RedisProvider struct {
	clientRequests chan *request
	redis          redis.Conn
}

type request struct {
	key     string
	value   string
	set     bool
	resultc chan *response
}

type response struct {
	value string
	ok    bool
}

func NewRedisProvider() *RedisProvider {
	provider := &RedisProvider{make(chan *request), nil}
	go provider.serve()
	return provider
}

func (p *RedisProvider) serve() {
	c, err := redis.Dial("tcp", ":6379")
	if err == nil {
		p.redis = c
		defer c.Close()
	} else {
		log.Printf("[cache] unable to connect to redis: %v\n", err)
	}
	for req := range p.clientRequests {
		if req.set {
			ok := p.set(req.key, req.value)
			req.resultc <- &response{"", ok}
		} else {
			value, ok := p.get(req.key)
			req.resultc <- &response{value, ok}
		}
	}
}

func (p *RedisProvider) get(key string) (value string, ok bool) {
	if p.redis == nil {
		return "", false
	}
	s, err := redis.String(p.redis.Do("GET", key))
	if err != nil {
		return "", false
	}
	return s, true
}

func (p *RedisProvider) set(key, value string) (ok bool) {
	if p.redis == nil {
		return false
	}
	_, err := p.redis.Do("SET", key, value)
	return err == nil
}

func (p *RedisProvider) ValueRequested(key string) (interface{}, bool) {
	request := &request{key, "", false, make(chan *response)}
	p.clientRequests <- request
	response := <-request.resultc
	return response.value, response.ok
}

func (p *RedisProvider) ValueArrived(key string, value interface{}) {
	s, ok := value.(string)
	if !ok {
		panic("redis provider only accepts string values")
	}
	request := &request{key, s, true, make(chan *response)}
	p.clientRequests <- request
	<-request.resultc
}
