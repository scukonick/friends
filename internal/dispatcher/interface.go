package dispatcher

import "context"

type Dispatcher interface {
	Connect(ctx context.Context, req *Request) (<-chan Response, error)
	Disconnect(ctx context.Context, req *Request) error
}
