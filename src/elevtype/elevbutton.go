package elevtype

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int        `json: "btneventFloor"`
	Button ButtonType `json: "btneventType"`
}

type ButtonLamp struct {
	Floor  int
	Button ButtonType
	Value  bool
}
