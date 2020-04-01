package bus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalBus(t *testing.T) {
	bus := NewLocalBus()

	testPublish := func(t *testing.T) {
		t.Parallel()

		ch, err := bus.Subscribe(context.Background(), []int{1, 2, 3}, 100)
		require.Nil(t, err)

		err = bus.Publish(context.Background(), "abc", []int{100})
		require.Nil(t, err)

		select {
		case val := <-ch:
			require.Equal(t, "abc", val)
		case <-time.After(10 * time.Millisecond):
			t.Error("did not get value from the channel")
		}
	}

	testBroadcast := func(t *testing.T) {
		t.Parallel()

		ch, err := bus.Subscribe(context.Background(), []int{10, 20, 30}, 200)
		require.Nil(t, err)

		err = bus.Broadcast(context.Background(), "xxx", 10)
		require.Nil(t, err)

		select {
		case val := <-ch:
			require.Equal(t, "xxx", val)
		case <-time.After(10 * time.Millisecond):
			t.Error("did not get value from the channel")
		}
	}

	testUnsubscribe := func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		ch, err := bus.Subscribe(ctx, []int{40, 50, 60}, 300)
		require.Nil(t, err)

		cancel()

		select {
		case _, ok := <-ch:
			require.False(t, ok) // channel should be closed
		case <-time.After(10000 * time.Millisecond):
			t.Error("did not get value from the channel")
		}
	}

	t.Run("publish", testPublish)
	t.Run("broadcast", testBroadcast)
	t.Run("unsubscribe", testUnsubscribe)
}
