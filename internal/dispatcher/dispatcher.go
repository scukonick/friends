package dispatcher

import (
	"context"

	"github.com/scukonick/friends/internal/bus"
)

type Dispatcher struct {
	b bus.Bus
}

func NewDispatcher(b bus.Bus) *Dispatcher {
	return &Dispatcher{
		b: b,
	}
}

func (d *Dispatcher) Connect(ctx context.Context, req *Request) (<-chan Response, error) {
	err := d.b.Broadcast(ctx, "online", req.UserID)
	if err != nil {
		return nil, err
	}

	in, err := d.b.Subscribe(ctx, req.Friends, req.UserID)
	if err != nil {
		return nil, err
	}

	out := make(chan Response)

	go func(out chan<- Response) {
		for m := range in {
			online := m == "online"
			out <- Response{
				Online: online,
			}
		}
	}(out)

	return out, nil
}

func (d *Dispatcher) Disconnect(ctx context.Context, req *Request) error {
	err := d.b.Publish(ctx, "offline", req.Friends)
	if err != nil {
		return err
	}

	return nil
}