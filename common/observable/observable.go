package observable

import (
	"errors"
	"sync"
)

type Observable struct {
	// 是一个只读的channel
	iterable Iterable
	// key 是只读的channel
	// value是无限大d的buffer
	listener map[Subscription]*Subscriber
	mux      sync.Mutex
	done     bool
}

// 观察者模式
func (o *Observable) process() {
	for item := range o.iterable {
		o.mux.Lock()
		for _, sub := range o.listener {
			// 向每个订阅者的buffer中写
			sub.Emit(item)
		}
		o.mux.Unlock()
	}
	o.close()
}

func (o *Observable) close() {
	o.mux.Lock()
	defer o.mux.Unlock()

	o.done = true
	for _, sub := range o.listener {
		// 每个订阅者的关掉
		sub.Close()
	}
}

// 订阅主题
func (o *Observable) Subscribe() (Subscription, error) {
	o.mux.Lock()
	defer o.mux.Unlock()
	if o.done {
		return nil, errors.New("Observable is closed")
	}
	subscriber := newSubscriber()
	o.listener[subscriber.Out()] = subscriber
	// 返回一个只读的channel
	return subscriber.Out(), nil
}

func (o *Observable) UnSubscribe(sub Subscription) {
	o.mux.Lock()
	defer o.mux.Unlock()
	subscriber, exist := o.listener[sub]
	if !exist {
		return
	}
	// 删除对应的订阅者
	delete(o.listener, sub)
	subscriber.Close()
}

func NewObservable(any Iterable) *Observable {
	observable := &Observable{
		iterable: any,
		listener: map[Subscription]*Subscriber{},
	}
	// 事件处理
	go observable.process()
	return observable
}
