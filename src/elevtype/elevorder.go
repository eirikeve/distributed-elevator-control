package elevtype

type OrderStatus int

const (
	Unknown    OrderStatus = 0
	Unassigned OrderStatus = 1
	Assigned   OrderStatus = 2
	Timeout    OrderStatus = 3
	Finished   OrderStatus = 4
)

type OrderInfo struct {
	S         OrderStatus
	timestamp int64
	assignee  string
}

type Elevorder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string
	// Floor & order type
	Order ButtonEvent
	// Timestamp at order creation [UNIX timestamp]
	TimestampCreations int64
	// Current status of order
	Status OrderStatus
	// Last update of order [UNIX timestamp]
	TimestampOrderChange int64
	// Assigned to elev @ IP
	Assignee string
}
