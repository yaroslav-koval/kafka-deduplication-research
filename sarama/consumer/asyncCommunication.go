package main

import "sync"

var (
	mtx          sync.Mutex
	messagesLeft int
)

var ConsCh = make(chan error)

func SetMessagesIndicator(value int) {
	mtx.Lock()
	messagesLeft = value
	mtx.Unlock()
}

func IncrementMessagesIndicator() {
	mtx.Lock()
	messagesLeft += 1
	mtx.Unlock()
}

func DecrementMessagesIndicator() {
	mtx.Lock()
	messagesLeft -= 1
	mtx.Unlock()
}
