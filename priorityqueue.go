// A priority queue built using the heap interface.
// https://pkg.go.dev/container/heap
package main

// A BuyBook implements heap.Interface and holds Orders.
type BuyBook []*Order

func (pq BuyBook) Len() int { return len(pq) }

func (pq BuyBook) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, price so we use greater than here.
	return pq[i].price > pq[j].price || (pq[i].price == pq[j].price && pq[i].timestamp < pq[j].timestamp)
}

func (pq BuyBook) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *BuyBook) Push(x any) {
	item := x.(*Order)
	*pq = append(*pq, item)
}

func (pq BuyBook) Peek() *Order {
	if len(pq) > 0 {
		return pq[0]
	}

	return nil
}

func (pq *BuyBook) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// A SellBook implements heap.Interface and holds Orders.
type SellBook []*Order

func (pq SellBook) Len() int { return len(pq) }

func (pq SellBook) Less(i, j int) bool {
	// We want Pop to give us the lowest, not highest, price so we use less than here.
	return pq[i].price < pq[j].price || (pq[i].price == pq[j].price && pq[i].timestamp < pq[j].timestamp)
}

func (pq SellBook) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *SellBook) Push(x any) {
	item := x.(*Order)
	*pq = append(*pq, item)
}

func (pq SellBook) Peek() *Order {
	if len(pq) > 0 {
		return pq[0]
	}

	return nil
}

func (pq *SellBook) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
