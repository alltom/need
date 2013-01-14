package need

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func ExampleNeed() {
	s := NewServer()
	value, err := s.Need("a key", func(string) (interface{}, error) {
		return 42, nil
	})
	if err == nil {
		if i, ok := value.(int); ok {
			fmt.Printf("%d\n", i)
		}
	}
	// Output: 42
}

// test the simplest case
func TestSimple(t *testing.T) {
	s := NewServer()
	value, err := s.Need("a key", func(string) (interface{}, error) {
		return 42, nil
	})
	if err == nil {
		if i, ok := value.(int); ok {
			if i == 42 {
				return // success!
			}
		}
	}
	t.FailNow()
}

// test the simplest case with 2 keys
func TestDoubleSimple(t *testing.T) {
	s := NewServer()
	value1, err1 := s.Need("one key", func(string) (interface{}, error) {
		return 42, nil
	})
	value2, err2 := s.Need("another key", func(string) (interface{}, error) {
		return 43, nil
	})
	if err1 == nil && err2 == nil {
		i1, ok1 := value1.(int)
		i2, ok2 := value2.(int)
		if ok1 && ok2 {
			if i1 == 42 && i2 == 43 {
				return // success!
			}
		}
	}
	t.FailNow()
}

// tests that an error propagates from the callback to Need's return value
func TestError(t *testing.T) {
	s := NewServer()
	_, err := s.Need("a key", func(string) (interface{}, error) {
		return nil, errors.New("bogus")
	})
	if err == nil || err.Error() != "bogus" {
		t.Fail()
	}
}

// tests that if one fails, the next one in the pool is consulted
func TestFailover(t *testing.T) {
	s := NewServer()
	s.Need("a key", func(string) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return nil, errors.New("bogus")
	})
	value, err := s.Need("a key", func(string) (interface{}, error) {
		return 42, nil
	})
	if err == nil {
		if i, ok := value.(int); ok {
			if i == 42 {
				return // success!
			}
		}
	}
	t.FailNow()
}

func TestMemoryProvider(t *testing.T) {
	s := NewServerWithProvider(NewMemoryProvider())
	value1, err1 := s.Need("a key", func(string) (interface{}, error) {
		return 42, nil
	})
	time.Sleep(10 * time.Millisecond)
	value2, err2 := s.Need("a key", func(string) (interface{}, error) {
		t.FailNow()
		panic("unreachable")
	})
	if err1 == nil && err2 == nil {
		i1, ok1 := value1.(int)
		i2, ok2 := value2.(int)
		if ok1 && ok2 {
			if i1 == 42 && i2 == 42 {
				return // success!
			}
		}
	}
	t.FailNow()
}
