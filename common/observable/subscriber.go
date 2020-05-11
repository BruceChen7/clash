package observable

import (
	"sync"

	"gopkg.in/eapache/channels.v1"
)

// 只读的
type Subscription <-chan interface{}

type Subscriber struct {
	// 无限大容量的channel
	buffer *channels.InfiniteChannel
	once   sync.Once
}

func (s *Subscriber) Emit(item interface{}) {
	// 返回一个可先的channel
	s.buffer.In() <- item
}

func (s *Subscriber) Out() Subscription {
	// 返回一个读Subscription
	return s.buffer.Out()
}

func (s *Subscriber) Close() {
	//  最多只执行一次
	s.once.Do(func() {
		s.buffer.Close()
	})
}

func newSubscriber() *Subscriber {
	sub := &Subscriber{
		buffer: channels.NewInfiniteChannel(),
	}
	return sub
}
