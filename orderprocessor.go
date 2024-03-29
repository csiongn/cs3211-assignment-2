package main

import (
	"container/heap"
	"context"
)

type OrderProcessor struct {
	orderChan    <-chan Order
	cancelChan   chan uint32
	orderMapping map[uint32]*Order
	buyBook      *BuyBook
	sellBook     *SellBook
}

func NewOrderProcessor(ctx context.Context, orderChan <-chan Order) {

	processor := &OrderProcessor{
		orderChan:    orderChan,
		cancelChan:   make(chan uint32, BufferSize),
		orderMapping: make(map[uint32]*Order),
		buyBook:      &BuyBook{},
		sellBook:     &SellBook{},
	}

	go processor.processOrders(ctx)
}

// processOrders listens for new orders and cancellation requests, directing them appropriately.
func (p *OrderProcessor) processOrders(ctx context.Context) {
	for {
		select {
		case order := <-p.orderChan:
			// Direct order processing based on its type.
			p.processOrder(order)
		case <-ctx.Done():
			// Context cancellation, terminate processing.
			return
		}
	}
}

func (p *OrderProcessor) processOrder(order Order) {
	// fmt.Fprintf(os.Stderr, "Processor[%s]: Received order id %d\n", order.instrument, order.orderId)
	switch order.inputType {
	case inputCancel:
		p.handleCancellation(order)
	case inputBuy:
		p.handleBuyOrder(order)
	case inputSell:
		p.handleSellOrder(order)
	}
}

func (p *OrderProcessor) handleCancellation(order Order) {
	if orderOriginal, ok := p.orderMapping[order.orderId]; ok {
		delete(p.orderMapping, order.orderId)
		orderOriginal.count = 0 // Set count of order to 0 as we can't immediately remove from priority queue
		OrderDeleted(order, true, GetCurrentTimestamp())
		order.done <- struct{}{}
	} else {
		// Order already processed or non-existent.
		// fmt.Fprintf(os.Stderr, "Processor[%s]: Unable to find order id %d in instrument map. rejecting cancel.\n", order.instrument, order.orderId)
		OrderDeleted(order, false, GetCurrentTimestamp())
		order.done <- struct{}{}
	}
}

func (p *OrderProcessor) handleBuyOrder(order Order) {
	for order.count > 0 {
		if p.sellBook.Len() > 0 {
			topOrder := p.sellBook.Peek()
			// fmt.Fprintf(os.Stderr, "Processor[%s]: Checked top sell order. Price is %d with count %d\n", order.instrument,
			// topOrder.price, topOrder.count)
			if topOrder.count == 0 {
				// Originally cancelled order
				heap.Pop(p.sellBook)
				continue
			}
			if order.price >= topOrder.price {
				// Match orders
				tradeQuantity := min(topOrder.count, order.count)
				order.count, topOrder.count = order.count-tradeQuantity, topOrder.count-tradeQuantity
				topOrder.executionId = topOrder.executionId + 1
				outputOrderExecuted(topOrder.orderId, order.orderId, topOrder.executionId, topOrder.price,
					tradeQuantity, GetCurrentTimestamp())

				if topOrder.count == 0 {
					delete(p.orderMapping, topOrder.orderId)
					heap.Pop(p.sellBook)
				}

				if order.count == 0 {
					order.done <- struct{}{}
				}
			} else {
				(&order).timestamp = GetCurrentTimestamp()
				heap.Push(p.buyBook, &order)
				p.orderMapping[order.orderId] = &order
				OrderAdded(order, GetCurrentTimestamp())
				order.done <- struct{}{}
				break
			}
		} else {
			(&order).timestamp = GetCurrentTimestamp()
			heap.Push(p.buyBook, &order)
			p.orderMapping[order.orderId] = &order
			OrderAdded(order, GetCurrentTimestamp())
			order.done <- struct{}{}
			break
		}
	}
}

func (p *OrderProcessor) handleSellOrder(order Order) {
	for order.count > 0 {
		if p.buyBook.Len() > 0 {
			topOrder := p.buyBook.Peek()
			// fmt.Fprintf(os.Stderr, "Processor[%s]: Checked top buy order. Price is %d with count %d\n", order.instrument,
			// topOrder.price, topOrder.count)
			if topOrder.count == 0 {
				// Originally cancelled order
				heap.Pop(p.buyBook)
				continue
			}
			if order.price <= topOrder.price {
				// Match orders
				tradeQuantity := min(topOrder.count, order.count)
				order.count, topOrder.count = order.count-tradeQuantity, topOrder.count-tradeQuantity
				topOrder.executionId = topOrder.executionId + 1
				outputOrderExecuted(topOrder.orderId, order.orderId, topOrder.executionId, topOrder.price,
					tradeQuantity, GetCurrentTimestamp())

				if topOrder.count == 0 {
					delete(p.orderMapping, topOrder.orderId)
					heap.Pop(p.buyBook)
				}

				if order.count == 0 {
					order.done <- struct{}{}
				}

			} else {
				(&order).timestamp = GetCurrentTimestamp()
				heap.Push(p.sellBook, &order)
				p.orderMapping[order.orderId] = &order
				OrderAdded(order, GetCurrentTimestamp())
				order.done <- struct{}{}
				break
			}
		} else {
			(&order).timestamp = GetCurrentTimestamp()
			heap.Push(p.sellBook, &order)
			p.orderMapping[order.orderId] = &order
			OrderAdded(order, GetCurrentTimestamp())
			order.done <- struct{}{}
			break
		}
	}
}
