package need_redis

import "fmt"

func ExampleBasic() {
	provider := NewRedisProvider()
	provider.ValueArrived("some key", "some value")
	if value, ok := provider.ValueRequested("some key"); ok {
		if s, ok := value.(string); ok {
			fmt.Printf("%s\n", s)
		}
	}
	// Output: some value
}
