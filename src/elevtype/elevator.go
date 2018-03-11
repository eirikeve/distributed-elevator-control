package elevtype

type ElevatorState int

const (
	Initializing ElevatorState = iota
	Idle         ElevatorState = iota
	Unloading                  = iota
	Moving                     = iota
)

type ElevatorErrorState int

const (
	FullFunctionality    ElevatorErrorState = iota
	ReducedFunctionality                    = iota
	NoFunctionality                         = iota
)

type Elevator struct {
	Floor               int
	MovementDirection   MotorDirection
	MovDirFromLastFloor MotorDirection 
	Orders              [NumFloors][NumButtons]SimpleOrder
	State               ElevatorState
	ErrorState          ElevatorErrorState
}
