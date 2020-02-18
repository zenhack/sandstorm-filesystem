package main

import (
	"sync"

	bridge_capnp "zenhack.net/go/sandstorm/capnp/sandstormhttpbridge"
)

type BridgePromise struct {
	ch     <-chan bridge_capnp.SandstormHttpBridge
	once   *sync.Once
	bridge bridge_capnp.SandstormHttpBridge
}

func NewBridgePromise() (*BridgePromise, chan<- bridge_capnp.SandstormHttpBridge) {
	ch := make(chan bridge_capnp.SandstormHttpBridge, 1)
	promise := &BridgePromise{ch: ch, once: &sync.Once{}}
	return promise, ch
}

func (p *BridgePromise) Wait() bridge_capnp.SandstormHttpBridge {
	p.once.Do(func() {
		p.bridge = <-p.ch
	})
	return p.bridge
}
