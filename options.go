package poller

// Option represents a configurable value passed to a Poller instantiation.
type Option func(*Poller) error

// SetInterval establishes an interval value for the poller. The value is
// represented in miliseconds.
func SetInterval(v int64) Option {
	return func(p *Poller) error {
		p.interval = v

		return nil
	}
}

// SetPusher appends a Pusher function to the pushers' slice. The pushers could
// be executed in parallel, so don't depend on the functions being executed in
// any given order. If you want to make something depend on them, then please
// make the dependency built-in within the pusher.
func SetPusher(fn Pusher) Option {
	return func(p *Poller) error {
		p.pushers = append(p.pushers, fn)

		return nil
	}
}

// SetOnError sets a function to be called if there's an error during the poll
// execution. This will replace the defaultOnError.
func SetOnError(fn OnError) Option {
	return func(p *Poller) error {
		p.onError = fn

		return nil
	}
}
