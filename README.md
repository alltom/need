need
====

I run a web service that processes users' Twitter timelines as tweets arrive. If multiple users subscribe to the same person, I used to naively process all of those tweets separately. With `need`, only the first user's thread processes the tweets; the others simply await the result.

    func ProcessTweet(tweetId int64) {
    	s := need.NewServer()
    	key := fmt.Sprintf("tweet-%d", tweetId)
    	value, err := s.Need(key, func(string) (interface{}, error) {
    		return expensiveProcessing(tweetId)
    	})
    	if err == nil {
    		if s, ok := value.(string); ok {
    			fmt.Printf("result: %s\n", s)
    		}
    	}
    }

How it works
------------

If multiple `Need`s arrive simultaneously with the same key, the callback for the first is called and its return value is passed to all the other callers.

If a callback returns an error, the next callback is tried. (The first caller gets the error.)

By default, `need` has no memory. If other requests for a key arrive after the original request completes, the new callback will be called to provide the value. You can give `need` a memory by supplying a `MemoryProvider` (uses an in-memory `map`) or `RedisProvider` (uses redis, though it panics for return values that aren't strings).

What it doesn't do (TODO)
-------------------------

`RedisProvider` should work with more than just strings. It should make some effort to marshal values if you want it to.

If you use `RedisProvider` (or any provider that works with multiple processes), you ought to be able to pool requests across them as well. Currently, `need` doesn't offer a rich enough API to providers to do that.
