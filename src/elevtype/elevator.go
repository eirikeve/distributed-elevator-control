package elevtype

const NumFloors = 4
const NumButtons = 3

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
	Orders              [NumFloors][NumButtons]ElevOrder
	State               ElevatorState
	ErrorState          ElevatorErrorState
}
