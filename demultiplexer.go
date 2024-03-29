package main

import (
	"context"
)

type demultiplexer struct {
	clientOrdersChan <-chan Order // Channel for incoming orders
	// cancelOrdersChan   chan uint32             // Channel for order cancellation IDs
	instrumentRouters  map[string]chan<- Order // Maps instrument symbols to their processing channels
	orderInstrumentMap map[uint32]string       // Maps order IDs to instrument symbols for cancellation
}

func Newdemultiplexer(orders <-chan Order, ctx context.Context) {
	demux := &demultiplexer{
		clientOrdersChan: orders,
		// cancelOrdersChan:   make(chan uint32, BufferSize),
		instrumentRouters:  make(map[string]chan<- Order),
		orderInstrumentMap: make(map[uint32]string),
	}

	go demux.run(ctx)
}

// run listens for incoming orders and cancellations, and routes them appropriately.
func (demux *demultiplexer) run(ctx context.Context) {
	for {
		select {
		case order := <-demux.clientOrdersChan: // Received new order
			// Handle new or cancellation orders

			if order.instrument == "" { // Cancellation order without instrument symbol
				if symbol, exists := demux.orderInstrumentMap[order.orderId]; exists { // Tries to retrieve order instrument
					order.instrument = symbol
				} else {
					// Unable to process cancellation; order ID not found
					// fmt.Fprintf(os.Stderr, "Demux: Unable to find order id %d in instrument map. rejecting cancel.\n", order.orderId)
					OrderDeleted(order, false, GetCurrentTimestamp())
					order.done <- struct{}{}
					continue
				}
			}

			// Register non-cancellation orders for potential future cancellations
			if order.inputType != inputCancel {
				demux.orderInstrumentMap[order.orderId] = order.instrument
			}

			// Route the order to the appropriate worker
			if router, exists := demux.instrumentRouters[order.instrument]; !exists {
				// Create new worker if none exists
				// fmt.Fprintf(os.Stderr, "Creating new order processor for %s\n", order.instrument)
				demux.spawnWorkerAndRoute(ctx, order)
			} else {
				router <- order
			}

		case <-ctx.Done():
			// Context cancelled; stop processing
			return
		}
	}
}

// Creates a worker for a new instrument and routes the order to it.
func (demux *demultiplexer) spawnWorkerAndRoute(ctx context.Context, order Order) {
	// Create a new channel for routing orders of this instrument
	instrumentChan := make(chan Order, BufferSize)
	demux.instrumentRouters[order.instrument] = instrumentChan

	// Spawn a new worker to process orders for this instrument
	NewOrderProcessor(ctx, instrumentChan)
	// fmt.Fprintf(os.Stderr, "Demux: Created new instrument channel for \n", order.instrument)
	// Route the current order to the new worker
	instrumentChan <- order
	// fmt.Fprintf(os.Stderr, "Demux: Sent order id %d to processor channel\n", order.orderId)
}
