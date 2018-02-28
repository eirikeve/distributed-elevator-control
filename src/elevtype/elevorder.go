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
	TimestampReceived float64 //[@todo]: Changed to float64 from int64, is this right?]
	// Current status of order
	Status OrderStatus
	// Last update of order [UNIX timestamp]
	TimestampLastOrderStatusChange float64 //[@todo]: Same as for TimestampReceived
	// Assigned to elev @ IP
	Assignee string
}

func EmptyOrder() ElevOrder {
	return ElevOrder{
		Id:                "",
		Order:             ButtonEvent{Floor: -1, Button: BT_HallUp},
		TimestampReceived: 0,
		Status:            Unknown,
		TimestampLastOrderStatusChange: 0,
		Assignee:                       "",
	}
}
