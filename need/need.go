package need

type Server struct {
	p       Provider
	c       chan *request
	waiters map[string][]*waiter
}

// may be called concurrently for different keys
// but calls for the same key are serialized
type Provider interface {
	// if ok is false, the needer's callback will be used to provide the value
	ValueRequested(key string) (value interface{}, ok bool)
	// called when one of the needer's callbacks provides a value for the key
	ValueArrived(key string, value interface{})
}

// somebody waiting on a value
type waiter struct {
	valuec    chan interface{}
	nonexistc chan chan interface{}
}

// if valuec is provided, taken to be a request for the value and nonexistc should also be provided
// if valuec is nil and valueFailed is false, value is taken to be the value for key
// if valuec is nil and valueFailed is true, taken to be a failed attempt to calculate the value
type request struct {
	key         string // required
	value       interface{}
	valueFailed bool
	valuec      chan interface{}      // chan to send the value to
	nonexistc   chan chan interface{} // chan to send request for value to
}

func (s *Server) receivedValue(key string, value interface{}) {
	waiters := s.waiters[key]
	delete(s.waiters, key)
	for _, w := range waiters {
		go func(w *waiter) {
			w.valuec <- value
		}(w)
	}
}

func (s *Server) valueCalculationFailed(key string) {
	if len(s.waiters[key]) > 0 {
		// pop a waiter
		w := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:len(s.waiters[key])]

		// ask them for the value
		returnValuec := make(chan interface{})
		go s.waitForValue(key, returnValuec)
		go func() { w.nonexistc <- returnValuec }()
	} else {
		// no one else was waiting for the value
		delete(s.waiters, key)
	}
}

func (s *Server) needValue(key string, valuec chan interface{}, nonexistc chan chan interface{}) {
	if waiters, ok := s.waiters[key]; ok {
		s.waiters[key] = append(waiters, &waiter{valuec, nonexistc})
	} else {
		s.waiters[key] = []*waiter{}

		go func() {
			if value, ok := s.p.ValueRequested(key); ok {
				valuec <- value
				s.c <- &request{key, value, false, nil, nil}
			} else {
				returnValuec := make(chan interface{})
				go s.waitForValue(key, returnValuec)
				nonexistc <- returnValuec
			}
		}()
	}
}

func (s *Server) waitForValue(key string, valuec chan interface{}) {
	value, ok := <-valuec
	if ok {
		s.p.ValueArrived(key, value)
		s.c <- &request{key, value, false, nil, nil}
	} else {
		s.c <- &request{key, nil, true, nil, nil}
	}
}

func (s *Server) serve() {
	for req := range s.c {
		if req.valuec == nil {
			if req.valueFailed {
				s.valueCalculationFailed(req.key)
			} else {
				s.receivedValue(req.key, req.value)
			}
		} else {
			s.needValue(req.key, req.valuec, req.nonexistc)
		}
	}
}

func (s *Server) Need(key string, f func(string) (interface{}, error)) (interface{}, error) {
	valuec := make(chan interface{})
	nonexistc := make(chan chan interface{})
	s.c <- &request{key, nil, false, valuec, nonexistc}
	select {
	case value := <-valuec:
		return value, nil
	case returnValuec := <-nonexistc:
		defer close(returnValuec)
		if value, err := f(key); err == nil {
			returnValuec <- value
			return value, nil
		} else {
			return value, err
		}
	}
	panic("unreachable")
}

func NewServer() *Server {
	return NewServerWithProvider(&NullProvider{})
}

func NewServerWithProvider(p Provider) *Server {
	server := &Server{p, make(chan *request), make(map[string][]*waiter)}
	go server.serve()
	return server
}
