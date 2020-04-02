package dispatcher_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/scukonick/friends/internal/dispatcher"
	"github.com/scukonick/friends/internal/generated/mocks"
)

func TestDispatcher(t *testing.T) {
	ctrl := gomock.NewController(t)

	bus := mocks.NewMockBus(ctrl)

	broadcastCall := bus.EXPECT().
		Broadcast(gomock.Any(), "online", 1).
		Return(nil).AnyTimes()

	connCall := bus.EXPECT().
		Subscribe(gomock.Any(), []int{2, 3}, 1).
		Return(make(chan string, 0), nil).
		After(broadcastCall)

	bus.EXPECT().
		Publish(gomock.Any(), "offline", []int{2, 3}).
		Return(nil).
		After(connCall)

	disp := dispatcher.NewBaseDispatcher(bus)

	connect := func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		req := &dispatcher.Request{
			UserID:  1,
			Friends: []int{2, 3},
		}
		_, err := disp.Connect(ctx, req)
		require.Nil(t, err)

		err = disp.Disconnect(ctx, req)
		require.Nil(t, err)
	}

	t.Run("connect", connect)
}
