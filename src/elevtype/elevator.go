package elevtype

const NFloors = 4
const NButtons = 3

type ElevatorAction int

const (
	Idle     ElevatorAction = iota
	DoorOpen                = iota
	Moving                  = iota
)

type ElevatorErrorState int

const (
	FullFunctionality    ElevatorErrorState = iota
	ReducedFunctionality                    = iota
	NoFunctionality                         = iota
)

type Elevator struct {
	Floor         int
	Direction     MotorDirection
	Orders        [NFloors][NButtons]ElevOrder
	CurrentAction ElevatorAction
	ErrorState    ElevatorErrorState
}
