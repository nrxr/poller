# poller

This is a simple and re-usable poller with modular and configurable parts.

```
                              ┌──────────────┐
                              │ p.onError()  │
        ┌─────────────────────│    called    │◀──────┐
        │                     └──────────────┘       │
        ▼                            ▲               │
┌──────────────┐      ┌──────────────┤      ┌─────────────────┐
│ ticker waits │      │  p.getter()  │      │ p.pushers slice │
│   interval   │─────▶│    called    │─────▶│    called in    │
└──────────────┘      └──────────────┘      │    sequence     │
        │                                   └─────────────────┘
        │                                            │
        │                                            │
        └────────────────────────────────────────────┘
```

Each of the elements listed in the diagram is configurable, you can set the
interval, pass a getter, pass several pushers and configure what's happening
when there's an error.

## Quick usage

```go
// main.go

func g(ctx context.Context) (interface{}, error) {
  // grab information from somewhere
  return "good", nil
}

func pu0(ctx context.Context, in interface{}) error {
  // push or transform information and do whatever
  log.Println(in.(string))
}

func main() {
  p, err := poller.New(g, poller.SetPusher(pu0), poller.SetInterval(5000))
  if err != nil {
    // do whatever with err...
  }

  p.Start(context.Background())
}
```
