package elevorderevaluation

import(
    "../elevtype"
    )

const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

/*
* Calculates how much time the Elevator will use to execute
* all its order, thus going into the Idle state
* @arg elev: Takes an Elevator as arguemtent, making it possible to simalute its actions
*/ 
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
            duration-=DOOR_OPEN_TIME/2
        default:
            //Should not be possible to enter default
            println("Entered defualt in timeToIdle. This should not happen")
	}
	for isSimulating == true{
        if requests_shouldStop(elev) == true{
            elev = requests_clearAtCurrentFloor(elev)
            duration += DOOR_OPEN_TIME;
            elev.Direction = requests_chooseDirection(elev);
            if(elev.Direction == elevtype.MD_Stop){
                return duration;
            }
        }
		elev.Floor += int(elev.Direction);
        duration += TRAVEL_TIME;                    //[POTENTIAL BUG] Not sure if converts motor type to int to floor
	}
    
    return duration
} 

/*
* Takes a list of Elevators, and simulates them all finding which elevator
* is best fit to take and execute an order.
* @arg elev[]: List of Elevators
*/
func delegateOrder(elevList []elevtype.Elevator) elevtype.Elevator {
    var durations[] int
    for _, elev :=range elevList{
        tempDuration:=timeToIdle(elev)
        durations = append(durations,tempDuration)
    }

    optElevIndex := findMaxIndex(durations)
    return elevList[optElevIndex]
} 

/*
* Takes a list of integers and returns the index
* containing the largest value
* @arg list: list containing integers 
*/
func findMaxIndex(list []int) int{
    
    var maxIndex int;
    var maxValue int;
    if list == nil{
        println("Empty list, Crashing")
    }
    for index,element :=range list{
        if (index == 0){
            maxValue = element
            maxIndex = index}
        if (element > maxValue){
            maxValue = element
            maxIndex = index
        }
    }
    return maxIndex
}
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

