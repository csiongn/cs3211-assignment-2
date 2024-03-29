package main

import "C"
import (
	"context"
	"net"
	"time"
)

const BufferSize = 100

type Engine struct {
	demultiplexerChannel chan<- Order
}

func NewEngine(ctx context.Context) *Engine {
	demultiplexerChannel := make(chan Order, BufferSize)
	Newdemultiplexer(demultiplexerChannel, ctx)
	engine := &Engine{demultiplexerChannel: demultiplexerChannel}
	return engine
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(conn, e.demultiplexerChannel)
}

func handleConn(conn net.Conn, demultiplexer chan<- Order) {
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

		demultiplexer <- order
		// fmt.Fprintf(os.Stderr, "Sent order id %d to mux\n", order.orderId)
		<-done
	}
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
