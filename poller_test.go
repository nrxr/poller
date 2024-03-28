package poller

import (
	"context"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name     string
		interval int64
	}{
		{
			name:     "default values",
			interval: 30000,
		},
		{
			name:     "set interval lower",
			interval: 15000,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(nil, SetInterval(tt.interval))
			if err != nil {
				t.Errorf("received error %v", err)
			}

			if p.interval != tt.interval {
				t.Errorf("interval %d is different to expected %d", p.interval, tt.interval)
			}
		})
	}
}

type errorCapturer struct {
	called bool
	err    error
}

func (ec *errorCapturer) onError(ctx context.Context, err error) {
	ec.called = true
	ec.err = err
}

type testingStruct struct {
	Member     []string
	TotalItems int64
}

func TestPoll_Poll(t *testing.T) {
	testPollErrWrongCount := errors.New("wrong TotalCount")

	cases := []struct {
		name     string
		getter   Getter
		pushers  []Pusher
		capturer *errorCapturer
		err      error
	}{
		{
			name: "simple getter, no pushers, no error",
			getter: func(context.Context) (interface{}, error) {
				return nil, nil
			},
			capturer: &errorCapturer{},
		},
		{
			name: "simple getter, simple pusher, no error",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{}, nil
			},
			pushers: []Pusher{
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if len(nn.Member) != 0 {
						return errors.New("not a testingStruct")
					}

					return nil
				},
			},
			capturer: &errorCapturer{},
		},
		{
			name: "simple getter, simple pusher, error wrong value set",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{TotalItems: 100}, nil
			},
			pushers: []Pusher{
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if nn.TotalItems != 0 {
						return testPollErrWrongCount
					}

					return nil
				},
			},
			capturer: &errorCapturer{},
			err:      testPollErrWrongCount,
		},
		{
			name: "simple getter, double pusher, error wrong value set in first",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{TotalItems: 100}, nil
			},
			pushers: []Pusher{
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if nn.TotalItems != 0 {
						return testPollErrWrongCount
					}

					return nil
				},
				func(ctx context.Context, in interface{}) error {
					return nil
				},
			},
			capturer: &errorCapturer{},
			err:      testPollErrWrongCount,
		},
		{
			name: "simple getter, double pusher, error wrong value set in second",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{TotalItems: 100}, nil
			},
			pushers: []Pusher{
				func(ctx context.Context, in interface{}) error {
					return nil
				},
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if nn.TotalItems != 0 {
						return testPollErrWrongCount
					}

					return nil
				},
			},
			capturer: &errorCapturer{},
			err:      testPollErrWrongCount,
		},
		{
			name: "simple getter, double pusher, error wrong value set in both",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{TotalItems: 100}, nil
			},
			pushers: []Pusher{
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if nn.TotalItems != 0 {
						return testPollErrWrongCount
					}

					return nil
				},
				func(_ context.Context, in interface{}) error {
					nn := in.(testingStruct)
					if nn.TotalItems != 0 {
						return testPollErrWrongCount
					}

					return nil
				},
			},
			capturer: &errorCapturer{},
			err:      testPollErrWrongCount,
		},
		{
			name: "simple getter, double pusher, no error",
			getter: func(context.Context) (interface{}, error) {
				return testingStruct{TotalItems: 100}, nil
			},
			pushers: []Pusher{
				func(ctx context.Context, in interface{}) error {
					return nil
				},
				func(ctx context.Context, in interface{}) error {
					return nil
				},
			},
			capturer: &errorCapturer{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{}

			opts = append(opts, SetOnError(tt.capturer.onError))
			for _, v := range tt.pushers {
				opts = append(opts, SetPusher(v))
			}

			nn, err := New(tt.getter, opts...)
			if err != nil {
				t.Errorf("on initialization: %v", err)
			}

			nn.Poll(context.Background())

			if tt.capturer.err != tt.err {
				t.Errorf("expected err %v and received %v", tt.err, tt.capturer.err)
			}
		})
	}
}
