package elevorderevaluation

import(
    et "../elevtype"
    fsm "../elevfsm"
    )

const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

/*
* Calculates how much time the Elevator will use to execute
* all its order, thus going into the Idle state
* @arg elev: Takes an Elevator as arguemtent, making it possible to simalute its actions
*/ 
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
            duration-=DOOR_OPEN_TIME/2
        default:
            //Should not be possible to enter default
            println("Entered defualt in timeToIdle. This should not happen")
	}
	for isSimulating == true{
        if fsm.OrderLogicCheckShouldStopAtFloor(elev) == true{
            elev = fsm.OrderLogicClearRequestsOnCurrentFloor(elev,elev.MovementDirection)
            duration += DOOR_OPEN_TIME;
            elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev);
            if(elev.MovementDirection == et.MD_Stop){
                return duration;
            }
        }
		elev.Floor += int(elev.MovementDirection);
        duration += TRAVEL_TIME;                    //[POTENTIAL BUG] Not sure if converts motor type to int to floor
	}
    
    return duration
} 

/*
* Takes a list of Elevators, and simulates them all finding which elevator
* is best fit to take and execute an order.
* @arg elev[]: List of Elevators
*/
func delegateOrder(elevList []et.Elevator) int {
    var durations[] int
    for _, elev :=range elevList{
        tempDuration:=timeToIdle(elev)
        durations = append(durations,tempDuration)
    }
    for _,element :=range durations{
        println(element)
    }

    optElevIndex := findMinIndex(durations)
    return optElevIndex
} 

/*
* Takes a list of integers and returns the index
* containing the largest value
* @arg list: list containing integers 
*/
func findMinIndex(list []int) int{
    
    var maxIndex int;
    var maxValue int;
    if list == nil{
        println("Empty list, Crashing")
    }
    for index,element :=range list{
        if (index == 0){
            maxValue = element
            maxIndex = index}
        if (element < maxValue){
            maxValue = element
            maxIndex = index
        }
    }
    return maxIndex
}
