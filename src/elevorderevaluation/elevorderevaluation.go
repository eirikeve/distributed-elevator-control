package elevorderevaluation

import(
<<<<<<< HEAD
    et "../elevtype"
    fsm "../elevfsm"
=======
    "../elevtype"
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
    )

const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

/*
* Calculates how much time the Elevator will use to execute
* all its order, thus going into the Idle state
* @arg elev: Takes an Elevator as arguemtent, making it possible to simalute its actions
*/ 
<<<<<<< HEAD
func timeToIdle(elev et.Elevator) int{
    duration := 0
	isSimulating := true

    switch(elev.State){
        case et.Idle:
            elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev);
            if elev.MovementDirection == et.MD_Stop{
                return duration
            }
        case et.Moving:
            duration += TRAVEL_TIME/2;
            elev.Floor += int(elev.MovementDirection);          //[POTENTIAL BUG] Not sure if converts motor type to int to floor
        case et.Unloading:                         //[@Todo]: Unloading is being changed to Unloading in master, must be changed when merginging
=======
func timeToIdle(elev elevtype.Elevator) int{
    duration := 0
	isSimulating := true

    switch(elev.CurrentAction){
        case elevtype.Idle:
            elev.Direction = requests_chooseDirection(elev);
            if elev.Direction == elevtype.MD_Stop{
                return duration
            }
        case elevtype.Moving:
            duration += TRAVEL_TIME/2;
            elev.Floor += int(elev.Direction);          //[POTENTIAL BUG] Not sure if converts motor type to int to floor
        case elevtype.DoorOpen:                         //[@Todo]: DoorOpen is being changed to Unloading in master, must be changed when merginging
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
            duration-=DOOR_OPEN_TIME/2
        default:
            //Should not be possible to enter default
            println("Entered defualt in timeToIdle. This should not happen")
	}
	for isSimulating == true{
<<<<<<< HEAD
        if fsm.OrderLogicCheckShouldStopAtFloor(elev) == true{
            elev = fsm.OrderLogicClearRequestsOnCurrentFloor(elev,elev.MovementDirection)
            duration += DOOR_OPEN_TIME;
            elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev);
            if(elev.MovementDirection == et.MD_Stop){
                return duration;
            }
        }
		elev.Floor += int(elev.MovementDirection);
=======
        if requests_shouldStop(elev) == true{
            elev = requests_clearAtCurrentFloor(elev)
            duration += DOOR_OPEN_TIME;
            elev.Direction = requests_chooseDirection(elev);
            if(elev.Direction == elevtype.MD_Stop){
                return duration;
            }
        }
		elev.Floor += int(elev.Direction);
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
        duration += TRAVEL_TIME;                    //[POTENTIAL BUG] Not sure if converts motor type to int to floor
	}
    
    return duration
} 

/*
* Takes a list of Elevators, and simulates them all finding which elevator
* is best fit to take and execute an order.
* @arg elev[]: List of Elevators
*/
<<<<<<< HEAD
func delegateOrder(elevList []et.Elevator) int {
=======
func delegateOrder(elevList []elevtype.Elevator) elevtype.Elevator {
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
    var durations[] int
    for _, elev :=range elevList{
        tempDuration:=timeToIdle(elev)
        durations = append(durations,tempDuration)
    }
<<<<<<< HEAD
    for _,element :=range durations{
        println(element)
    }

    optElevIndex := findMinIndex(durations)
    return optElevIndex
=======

    optElevIndex := findMaxIndex(durations)
    return elevList[optElevIndex]
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
} 

/*
* Takes a list of integers and returns the index
* containing the largest value
* @arg list: list containing integers 
*/
<<<<<<< HEAD
func findMinIndex(list []int) int{
=======
func findMaxIndex(list []int) int{
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
    
    var maxIndex int;
    var maxValue int;
    if list == nil{
        println("Empty list, Crashing")
    }
    for index,element :=range list{
        if (index == 0){
            maxValue = element
            maxIndex = index}
<<<<<<< HEAD
        if (element < maxValue){
=======
        if (element > maxValue){
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
            maxValue = element
            maxIndex = index
        }
    }
    return maxIndex
}
<<<<<<< HEAD
=======
//[@Todo]: Remove these functions. Only made to check if other functions in this folder are correctly implemented 
func requests_chooseDirection(elev elevtype.Elevator)elevtype.MotorDirection{
    //Fictive function, does nothing
    return elevtype.MD_Stop
}

func requests_shouldStop(elev elevtype.Elevator) bool{
    //Fictive function, does nothing
    return true
}

func requests_clearAtCurrentFloor(elev elevtype.Elevator) elevtype.Elevator{
    //Fictive function, does nothing
    return elev
}

//random comment
>>>>>>> 9fd3eb0fe80f9f7a0a07284664ef1f6281e0b695
