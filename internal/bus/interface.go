package bus

import "context"

/*

4. When another connection established with the user_id from the list of any other user's "friends" section,
they should be notified about it with message {"online": true}

5. When the user goes offline, his "friends"
(if it has any and any of them online) should receive a message {"online": false
*/

type Bus interface {
	Publish(ctx context.Context, text string, destIDs []int) error
	Broadcast(ctx context.Context, text string, sourceID int) error
	Subscribe(ctx context.Context, sourceIDs []int, destID int) (<-chan string, error)
}
