package elevorderevaluation

import(
    "../elevtype"
)
/*int timeToIdle(Elevator elev){
    int duration = 0;
    
    switch(elev.behaviour){
    case EB_Idle:
        elev.dirn = requests_chooseDirection(elev);
        if(elev.dirn == D_Stop){
            return duration;
        }
        break;
    case EB_Moving:
        duration += TRAVEL_TIME/2;
        elev.floor += elev.dirn;
        break;
    case EB_DoorOpen:
        duration -= DOOR_OPEN_TIME/2;
    }


    while(true){
        if(requests_shouldStop(elev)){
            elev = requests_clearAtCurrentFloor(elev, NULL);
            duration += DOOR_OPEN_TIME;
            elev.dirn = requests_chooseDirection(elev);
            if(elev.dirn == D_Stop){
                return duration;
            }
        }
        elev.floor += elev.direction;
        duration += TRAVEL_TIME;
    }
    
}*/

const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

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

//[@Todo]: Remove these functions. Only made to check if other functions in this folder are correctly implemented 
func requests_chooseDirection(elev elevtype.Elevator)elevtype.MotorDirection{
    //Fictive function, does nothing
}

func requests_shouldStop(elev elevtype.Elevator) bool{
    //Fictive function, does nothing
}

func requests_clearAtCurrentFloor(elev elevtype.Elevator) elevtype.Elevator{
    //Fictive function, does nothing
}