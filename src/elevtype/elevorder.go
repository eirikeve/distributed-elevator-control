package elevtype

import "time"

// OrderStatus is the current state of an elevator order
type OrderStatus int

const (
	Unknown  OrderStatus = 0
	Received OrderStatus = 1
	Accepted OrderStatus = 2
	Finished OrderStatus = 3
)

// https://stackoverflow.com/questions/35660467/how-to-access-fields-of-a-json-in-go
type ElevOrder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string `json: "orderID"`
	// Floor & order type
	Order ButtonEvent `json: "orderBtnEv"`
	// Timestamp at order creation [UNIX timestamp]
	TimestampReceived int64 `json: "orderLastUpdate"`
	// Current status of order
	Status OrderStatus `json: "orderStatus"`
	// Last update of order [UNIX timestamp]
	TimestampLastOrderStatusChange int64 `json: "orderLastStatusChange"`
	// Assigned to elev w/ this ID
	Assignee int32 `json: "orderAssignee"`
	// List over IDs of elevators which have acknowledged this order
	Acks []int32 `json: "Acks"`
	// Marks whether the order has been sent to the Assignee local elevator queue (to the FSM queue)
	SentToAssigneeElevator bool `json: "sent"`
}


// A simpler order representation used by the local system, to the elevator
type SimpleOrder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string
	// Floor & order type
	Order          ButtonEvent
	TagRemoveOrder bool
}

type GeneralOrder interface {
	ToSimpleOrder() SimpleOrder
	IsEmpty() bool
	IsActive() bool
	IsAccepted() bool
	IsSame(other GeneralOrder) bool
	GetID() string
	GetFloor() int
	GetButton() ButtonType
	GetOrder() ButtonEvent
	IsCabOrder() bool
	TimeSinceTimeout() int64
	TagRemove() bool
}

func (o SimpleOrder) ToSimpleOrder() SimpleOrder {
	return o
}
func (o ElevOrder) ToSimpleOrder() SimpleOrder {
	return SimpleOrder{Id: o.Id, Order: o.Order}
}

func (o SimpleOrder) IsEmpty() bool {
	return o.Id == ""
}
func (o ElevOrder) IsEmpty() bool {
	return o.Id == ""
}

func (o SimpleOrder) IsActive() bool {
	return o.Id != ""
}
func (o ElevOrder) IsActive() bool {
	return o.Id != "" && (o.Status == Accepted || o.Status == Received)
}
func (o SimpleOrder) IsAccepted() bool {
	return o.Id != ""
}
func (o ElevOrder) IsAccepted() bool {
	return o.Id != "" && o.Status == Accepted
}
func (o SimpleOrder) IsSame(other GeneralOrder) bool {
	return o.GetID() == other.GetID()
}
func (o ElevOrder) IsSame(other GeneralOrder) bool {
	return o.GetID() == other.GetID()
}
func (o SimpleOrder) GetID() string {
	return o.Id
}
func (o ElevOrder) GetID() string {
	return o.Id
}
func (o SimpleOrder) GetFloor() int         { return o.Order.Floor }
func (o ElevOrder) GetFloor() int           { return o.Order.Floor }
func (o SimpleOrder) GetButton() ButtonType { return o.Order.Button }
func (o ElevOrder) GetButton() ButtonType   { return o.Order.Button }
func (o SimpleOrder) GetOrder() ButtonEvent { return o.Order }
func (o ElevOrder) GetOrder() ButtonEvent   { return o.Order }
func (o ElevOrder) IsCabOrder() bool        { return o.Order.Button == BT_Cab }
func (o SimpleOrder) IsCabOrder() bool      { return o.Order.Button == BT_Cab }
func (o ElevOrder) TimeSinceTimeout() int64 {
	if !o.IsEmpty() {
		return time.Now().Unix() - (o.TimestampReceived + OrderTimeoutSeconds)
	}
	return -1

}
func (o SimpleOrder) TimeSinceTimeout() int64 {
	return -1
}
func (o ElevOrder) TagRemove() bool   { return false }
func (o SimpleOrder) TagRemove() bool { return o.TagRemoveOrder }

func EmptyOrder() ElevOrder {
	return ElevOrder{
		Id:                "",
		Order:             ButtonEvent{Floor: -1, Button: BT_HallUp},
		TimestampReceived: 0,
		Status:            Unknown,
		TimestampLastOrderStatusChange: 0,
		Assignee:                       -1,
		SentToAssigneeElevator:         false,
	}
}
