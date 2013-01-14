package need

// doesn't cache anything
type NullProvider struct {
}

func (p *NullProvider) ValueRequested(key string) (interface{}, bool) {
	return nil, false
}

func (p *NullProvider) ValueArrived(key string, value interface{}) {
}
