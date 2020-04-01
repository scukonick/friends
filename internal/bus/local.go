package bus

import (
	"context"
	"errors"
	"log"
	"sync"
)

const maxQueueLen = 1024

type msg struct {
	text string
}

func newMsg(text string) msg {
	return msg{
		text: text,
	}
}

type set map[int]struct{}

type subs map[int]set

func newSubs() subs {
	return make(map[int]set)
}

func (s subs) getSubscribedIDs(sourceID int) set {
	return s[sourceID]
}

func (s subs) addSubscription(from, to int) {
	if _, ok := s[from]; !ok {
		s[from] = make(set)
	}

	s[from][to] = struct{}{}
}

func (s subs) cancelSubscription(to int) {
	for from := range s {
		delete(s[from], to)
	}
}

type LocalBus struct {
	queues        map[int]chan msg
	subscriptions subs
	lock          sync.RWMutex
}

func NewLocalBus() *LocalBus {
	return &LocalBus{
		queues:        make(map[int]chan msg),
		subscriptions: newSubs(),
		lock:          sync.RWMutex{},
	}
}

func (l *LocalBus) Publish(ctx context.Context, text string, destIDs []int) error {
	msg := newMsg(text)

	l.lock.RLock()
	defer l.lock.RUnlock()

	dest := make(set, len(destIDs))
	for _, id := range destIDs {
		dest[id] = struct{}{}
	}

	l.unsafeSendToIDs(msg, dest)

	// always nil but in other implementations may be different
	return nil
}

func (l *LocalBus) Broadcast(ctx context.Context, text string, sourceID int) error {
	m := newMsg(text)

	l.lock.RLock()
	defer l.lock.RUnlock()

	l.unsafeSendToIDs(m, l.subscriptions.getSubscribedIDs(sourceID))

	return nil
}

func (l *LocalBus) unsafeSendToIDs(m msg, ids set) {
	for id := range ids {
		q, ok := l.queues[id]
		if !ok {
			continue
		}

		select {
		// sending msg until out channel is full
		case q <- m:
			// do nothing
		default:
			log.Printf("WARN: chan for user %d is full\n", id)
			continue
		}
	}
}

func (l *LocalBus) Subscribe(ctx context.Context, sourceIDs []int, destID int) (<-chan string, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.queues[destID]; ok {
		return nil, errors.New("duplicated subscription is not supported yet")
	}

	innerChan := make(chan msg, maxQueueLen)
	l.queues[destID] = innerChan

	for _, sourceID := range sourceIDs {
		l.subscriptions.addSubscription(sourceID, destID)
	}

	outChan := make(chan string)

	go func(out chan<- string) {
		defer l.dropSubscriptions(destID)
		defer close(outChan)

		for {
			select {
			case <-ctx.Done():
				return
			case m := <-innerChan:
				outChan <- m.text
			}
		}
	}(outChan)

	return outChan, nil
}

func (l *LocalBus) dropSubscriptions(id int) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.subscriptions.cancelSubscription(id)
	delete(l.queues, id)
}
