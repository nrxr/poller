// Package poller provides logic for an abstract and re-usable poller.
//
// This package serves as a simple template where you can throw a getter and a
// pusher (or multiple) and get an easily modular and configurable poller
// running in no-time, worrying only about your logic and not caring about a
// pattern you've been using over and over and don't want to write anymore in
// your life. Ever.
//
//	                              ┌──────────────┐
//	                              │ p.onError()  │
//	        ┌─────────────────────│    called    │◀──────┐
//	        │                     └──────────────┘       │
//	        ▼                            ▲               │
//	┌──────────────┐      ┌──────────────┤      ┌─────────────────┐
//	│ ticker waits │      │  p.getter()  │      │ p.pushers slice │
//	│   interval   │─────▶│    called    │─────▶│    called in    │
//	└──────────────┘      └──────────────┘      │    sequence     │
//	        │                                   └─────────────────┘
//	        │                                            │
//	        │                                            │
//	        └────────────────────────────────────────────┘
package poller

import (
	"context"
	"log"
	"time"
)

// Poller represents a polling instance. Is initialized via New and
// configurable with Option(s).
type Poller struct {
	interval int64
	getter   Getter
	pushers  []Pusher
	onError  OnError
}

// Getter returns a value and an error. Getter is used as a template for a
// getter function passed to a Poller initialization. The Poller instance will
// use Getter to update the status of the Poller.
type Getter func(context.Context) (interface{}, error)

// Pusher returns an error. Pusher is used as a template for functions to be
// executed after the getter updates the status in the Poller instance. The
// interface value passed as input is the same interface{} from the output.
type Pusher func(context.Context, interface{}) error

// OnError represents the function to be called by p.Poll's method after an
// error occurs on the call of the Getter or any of the Pushers. The arguments
// passed are the context and the error returned by either the Pusher or the
// Getter. By default it'll call log.Println on the error value. You can inject
// dependencies by using a method from an object with the injected
// dependencies, then you can send the error to whatever you want or need
// without modifying the Poller object or injecting a dependency directly.
type OnError func(context.Context, error)

func defaultOnError(_ context.Context, err error) {
	log.Println(err)
}

// New creates a Poller instance with the default values, modified with the
// Option values passed. It'll return an error
func New(g Getter, opts ...Option) (Poller, error) {
	p := Poller{ // default values
		interval: 30000, // 30s
		getter:   g,
		onError:  defaultOnError,
	}

	for _, opt := range opts {
		if err := opt(&p); err != nil {
			return p, err
		}
	}

	return p, nil
}

// Start begins the polling mechanism in the set interval. This is a blocking
// call. Use the context passed to cancel this call.
func (p Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.interval) * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			go p.Poll(ctx)
		}
	}
}

// Poller calls p's getter and pushers, if the getter succeeds. Each pusher is
// called independently and if one pusher errors out it wont cancel the other
// one. All the pushers are called. On error, p.onError will be called for
// both, the getter and the pushers.
func (p Poller) Poll(ctx context.Context) {
	gr, err := p.getter(ctx)
	if err != nil {
		p.onError(ctx, err)
		return
	}

	for _, pp := range p.pushers {
		err := pp(ctx, gr)
		if err != nil {
			p.onError(ctx, err)
		}
	}
}
