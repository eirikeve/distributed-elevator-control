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

// https://stackoverflow.com/questions/35660467/how-to-access-fields-of-a-json-in-go
type ElevOrder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string `json: "orderID"`
	// Floor & order type
	Order ButtonEvent `json: "orderBtnEv"`
	// Timestamp at order creation [UNIX timestamp]
	TimestampReceived int64 `json: "orderLastUpdate"`
	//[@todo]: Changed to float64 from int64, is this right?]
	// Current status of order
	Status OrderStatus `json: "orderStatus"`
	// Last update of order [UNIX timestamp]
	TimestampLastOrderStatusChange int64 `json: "orderLastStatusChange"`
	//[@todo]: Same as for TimestampReceived
	// Assigned to elev @ IP
	Assignee string `json: "orderAssignee"`
}

type SimpleOrder struct {
	// ID of request IP+timestamp+ButtonEvent
	Id string
	// Floor & order type
	Order ButtonEvent
}

type GeneralOrder interface {
	ToSimpleOrder() SimpleOrder
	//IsSimpleOrder() bool
	IsEmpty() bool
	IsActive() bool
	IsSame(other GeneralOrder) bool
	GetID() string
	GetFloor() int
	GetButton() ButtonType
	GetOrder() ButtonEvent
}

func (o SimpleOrder) ToSimpleOrder() SimpleOrder {
	return o
}
func (o ElevOrder) ToSimpleOrder() SimpleOrder {
	return SimpleOrder{Id: o.Id, Order: o.Order}
}

//func (o SimpleOrder) IsSimpleOrder() bool {
//	return true
//}
//func (o ElevOrder) IsSimpleOrder() bool {
//	return false
//}
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
