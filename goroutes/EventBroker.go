package goroutes

// 发布订阅模型

type Event struct {
	Job       *Job
	EventType string
}

type eventBroker struct {
	stop        chan struct{}
	publish     chan Event
	subscribe   chan chan Event
	unsubscribe chan chan Event
	send        chan Event
}

// 开始一个持续的事件监听

func (broker *eventBroker) Start() {
	subscribers := map[chan Event]struct{}{}
	for {
		select {
		case <-broker.stop: // 获取一个停止事件消息
			for sub := range subscribers {
				close(sub) // 关闭每一个事件
			}
			return
		case sub := <-broker.subscribe: // 订阅消息
			subscribers[sub] = struct{}{}
		case sub := <-broker.unsubscribe: // 个退订消息
			delete(subscribers, sub)
		case event := <-broker.publish: // 发布事件消息
			for sub := range subscribers {
				sub <- event
			}
		}
	}
}

// 停止消息订阅

func (broker *eventBroker) Stop() {
	close(broker.stop)
}

// 订阅消息

func (broker *eventBroker) Subscribe() chan Event {
	events := make(chan Event, 5)
	broker.subscribe <- events // 将监听到的事件写入subscribe信道
	return events
}

// 取消订阅

func (broker *eventBroker) Unsubscribe(events chan Event) {
	broker.unsubscribe <- events
	close(events)
}

// 发布订阅

func (broker *eventBroker) Publish(event Event) {
	broker.publish <- event
}

func newBroker() *eventBroker {
	broker := &eventBroker{
		stop:        make(chan struct{}),
		publish:     make(chan Event, 5),
		subscribe:   make(chan chan Event, 5),
		unsubscribe: make(chan chan Event, 5),
		send:        make(chan Event, 5),
	}
	go broker.Start()
	return broker
}

var EventBroker = newBroker()
