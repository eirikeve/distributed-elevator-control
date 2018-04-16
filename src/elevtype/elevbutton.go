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

/*Check if Button event is a CAB button
 */
func IsCabButton(b ButtonEvent) bool {
	if b.Button == BT_Cab {
		return true
	} else {
		return false
	}
}
