package main

// Order represents an order in the matching engine.
type Order struct {
	orderId     uint32
	executionId uint32
	instrument  string
	price       uint32
	count       uint32
	timestamp   int64
	inputType   inputType
	done        chan<- struct{} // to mark that the current order has been processed and the client thread can send more orders
}

// NewOrder creates a new Order.
func NewOrder(orderID, executionId, price, count uint32, instrument string, inputType inputType, done chan<- struct{}) Order {
	return Order{
		orderId:     orderID,
		executionId: executionId,
		instrument:  instrument,
		price:       price,
		count:       count,
		inputType:   inputType,
		done:        done,
	}
}
