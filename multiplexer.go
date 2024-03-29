package main

import (
	"context"
)

type Multiplexer struct {
	clientOrdersChan <-chan Order // Channel for incoming orders
	// cancelOrdersChan   chan uint32             // Channel for order cancellation IDs
	instrumentRouters  map[string]chan<- Order // Maps instrument symbols to their processing channels
	orderInstrumentMap map[uint32]string       // Maps order IDs to instrument symbols for cancellation
}

func NewMultiplexer(orders <-chan Order, ctx context.Context) {
	mux := &Multiplexer{
		clientOrdersChan: orders,
		// cancelOrdersChan:   make(chan uint32, BufferSize),
		instrumentRouters:  make(map[string]chan<- Order),
		orderInstrumentMap: make(map[uint32]string),
	}

	go mux.run(ctx)
}

// run listens for incoming orders and cancellations, and routes them appropriately.
func (mux *Multiplexer) run(ctx context.Context) {
	for {
		select {
		case order := <-mux.clientOrdersChan: // Received new order
			// Handle new or cancellation orders

			if order.instrument == "" { // Cancellation order without instrument symbol
				if symbol, exists := mux.orderInstrumentMap[order.orderId]; exists { // Tries to retrieve order instrument
					order.instrument = symbol
				} else {
					// Unable to process cancellation; order ID not found
					// fmt.Fprintf(os.Stderr, "Mux: Unable to find order id %d in instrument map. rejecting cancel.\n", order.orderId)
					OrderDeleted(order, false, GetCurrentTimestamp())
					order.done <- struct{}{}
					continue
				}
			}

			// Register non-cancellation orders for potential future cancellations
			if order.inputType != inputCancel {
				mux.orderInstrumentMap[order.orderId] = order.instrument
			}

			// Route the order to the appropriate worker
			if router, exists := mux.instrumentRouters[order.instrument]; !exists {
				// Create new worker if none exists
				// fmt.Fprintf(os.Stderr, "Creating new order processor for %s\n", order.instrument)
				mux.spawnWorkerAndRoute(ctx, order)
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
func (mux *Multiplexer) spawnWorkerAndRoute(ctx context.Context, order Order) {
	// Create a new channel for routing orders of this instrument
	instrumentChan := make(chan Order, BufferSize)
	mux.instrumentRouters[order.instrument] = instrumentChan

	// Spawn a new worker to process orders for this instrument
	NewOrderProcessor(ctx, instrumentChan)
	// fmt.Fprintf(os.Stderr, "Mux: Created new instrument channel for \n", order.instrument)
	// Route the current order to the new worker
	instrumentChan <- order
	// fmt.Fprintf(os.Stderr, "Mux: Sent order id %d to processor channel\n", order.orderId)
}
