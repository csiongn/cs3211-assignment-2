package main

import "C"
import (
	"context"
	"net"
	"time"
)

const BufferSize = 100

type Engine struct {
	multiplexerChannel chan<- Order
}

func NewEngine(ctx context.Context) *Engine {
	multiplexerChannel := make(chan Order, BufferSize)
	NewMultiplexer(multiplexerChannel, ctx)
	engine := &Engine{multiplexerChannel: multiplexerChannel}
	return engine
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(conn, e.multiplexerChannel)
}

func handleConn(conn net.Conn, multiplexer chan<- Order) {
	defer conn.Close()
	done := make(chan struct{})
	for {
		in, err := readInput(conn)
		if err != nil {
			return
		}

		// Create new order object
		order := NewOrder(in.orderId, 0, in.price, in.count, in.instrument,
			in.orderType, done)

		multiplexer <- order
		// fmt.Fprintf(os.Stderr, "Sent order id %d to mux\n", order.orderId)
		<-done
	}
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
