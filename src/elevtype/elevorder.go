package elevtype

// OrderStatus is the current state of an elevator order
type OrderStatus int

const (
	Timeout  OrderStatus = -2
	Unknown  OrderStatus = -1
	Received OrderStatus = 0
	Accepted OrderStatus = 1
	Finished OrderStatus = 2
)

type ElevOrder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string
	// Floor & order type
	Order ButtonEvent
	// Timestamp at order creation [UNIX timestamp]
	TimestampReceived int64
	// Current status of order
	Status OrderStatus
	// Last update of order [UNIX timestamp]
	TimestampLastOrderStatusChange int64
	// Assigned to elev @ IP
	Assignee string
}
