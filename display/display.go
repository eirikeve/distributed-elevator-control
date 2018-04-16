package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	b "../src/elevnetwork/bcast"
	locip "../src/elevnetwork/localip"
	et "../src/elevtype"
	tm "github.com/buger/goterm"
)

/*

display is a separate program from the main project.
It is not necessary to run display to run the main project.

display monitors the network for UDP messages from any active systems,
these regular updates are stored, and displayed in a HUD.
Due to the lib used, you may not see anything you write in the terminal after exiting the program. Haven't found a fix.

Since display is not a part of the project and will likely never be used again, we haven't documented it as well as the main project.

*/

func main() {
	LocIP, _ := locip.LocalIP()
	var recvExitSignal = make(chan bool, 2)
	var regularUpdateTimer = time.Now()
	var regularUpdateTimeout = time.Second * 60
	var systems []et.ElevState

	// Can press p to exit.
	go exitPoller(recvExitSignal)

	// Start Transmitter and Receiver for sending messages
	var sendRegularUpdates = make(chan et.ElevState, 12)
	var recvRegularUpdates = make(chan et.ElevState, 12)

	// Used to display how many messages we receive, and which are not empty
	var countReceivedMsg int64
	var countReceivedMsgUsable int64

	go b.Transmitter(et.AckHandlerPort, sendRegularUpdates)
	go b.Receiver(et.AckHandlerPort, recvRegularUpdates)

	// Necessary to have empty screen before drawing
	tm.Clear()

	// If we don't receive a single message for a set time, close the program
	for time.Now().Sub(regularUpdateTimer) < regularUpdateTimeout {
		select {
		case <-recvExitSignal:
			tm.Clear()
			println("Received stop signal. Exiting Display.")
			return
		// Received a regular update message, so store the system
		case remoteUpdate := <-recvRegularUpdates:
			countReceivedMsg++
			if remoteUpdate.ID != 0 {
				countReceivedMsgUsable++
				regularUpdateTimer = time.Now()
				existsInSystems := false
				index := -1
				for i, sys := range systems {
					if sys.ID == remoteUpdate.ID {
						existsInSystems = true
						index = i
						break
					}
				}
				if existsInSystems {
					systems[index] = remoteUpdate
				} else {
					systems = append(systems, remoteUpdate)
				}
			}
		default:
		}

		// Drawing the systems
		tm.MoveCursor(1, 1)
		tm.Println("Elevator Display Tool\nLocal IP:  \t" + LocIP + "\nActive systems:\t" + strconv.FormatInt(int64(len(systems)), 10) + "\nTimeout in:  \t" + strconv.FormatInt(60-int64(time.Now().Sub(regularUpdateTimer).Seconds()), 10) + "s" + "\nTotal recv messages: " + strconv.FormatInt(countReceivedMsg, 10) + " (non-empty: " + strconv.FormatInt(countReceivedMsgUsable, 10) + ")")

		minWidth := 0
		tabWidth := 10
		padding := 0
		padChar := byte(' ')
		if len(systems) > 0 {
			for i, sys := range systems {
				tm.Println("============================================-")
				tm.Println(strconv.FormatInt(int64(sys.ID), 10) + " (" + strconv.FormatInt(int64(i), 10) + ")                 " + "                  |")
				table := tm.NewTable(minWidth, tabWidth, padding, padChar, 0)

				fmt.Fprintf(table, "State\t \t|\tOrders\t \t \tDone:\t  "+intToBufferedString(int64(len(sys.FinishedOrders)), 4)+"\t|\n")
				fmt.Fprintf(table, "---------\t-----\t|\t---------\t-----\t-----\t-----\t-----\t|\n")
				fmt.Fprintf(table, "Floor \t%d\t|\tFloor\t0\t1\t2\t3\t|\n", sys.E.Floor)
				fmt.Fprintf(table, "State \t%d\t|\tHallUp\t"+orderToString(&sys, 0, 0)+"\t"+orderToString(&sys, 1, 0)+"\t"+orderToString(&sys, 2, 0)+"\t"+orderToString(&sys, 3, 0)+"\t|\n", sys.E.State)
				fmt.Fprintf(table, "ErrState \t%d\t|\tHallDown \t"+orderToString(&sys, 0, 1)+"\t"+orderToString(&sys, 1, 1)+"\t"+orderToString(&sys, 2, 1)+"\t"+orderToString(&sys, 3, 1)+"\t|\n", sys.E.ErrorState)
				fmt.Fprintf(table, "MovDir \t%d\t|\tCab\t"+orderToString(&sys, 0, 2)+"\t"+orderToString(&sys, 1, 2)+"\t"+orderToString(&sys, 2, 2)+"\t"+orderToString(&sys, 3, 2)+"\t|\n", sys.E.MovementDirection)
				fmt.Fprintf(table, elevFloorDisplay(sys.E.Floor, sys.E.MovementDirection)+"\n")
				tm.Print(table)
			}
			tm.Println("                                            |")
			tm.Println("============================================-")

		}
		// Render
		tm.Flush()
	}
	tm.Clear()
	println("Timed out. Exiting Display")

}

//elevFloorDisplay gives a string of spaces (ending with |) where the position of an E (and an arrow) is determined by  floor & dir
func elevFloorDisplay(floor int, movDir et.MotorDirection) string {
	var s = "                  "

	for f := 0; f < floor+1; f++ {
		s += "     "
	}
	switch movDir {
	case et.MD_Down:
		s += "<E "
	case et.MD_Stop:
		s += " E "
	case et.MD_Up:
		s += " E>"
	default:
		s += " E "
	}
	for f := floor + 1; f < et.NumFloors; f++ {
		s += "     "
	}
	s += "   |"
	return s
}

//intToGenericString returns a string representing the number if it is 0-9, else it returns ">9"
func intToGenericString(i int64) string {
	if i == 0 {
		return " 0"
	} else if i <= 9 {
		return " " + strconv.FormatInt(i, 10)
	} else {
		return ">9"
	}
}

//intToBufferedString gives a string with the number i and then some padding untill length is sz
func intToBufferedString(i int64, sz int) string {
	s := strconv.FormatInt(i, 10)
	if len(s) < sz {
		for i := 0; i < (sz - len(s)); i++ {
			s = " " + s
		}
	}
	return s
}

// orderToString gives a string representation of an order
func orderToString(sys *et.ElevState, f int, b int) string {
	orderState := et.Unknown
	if sys.CurrentOrders[f][b].Id != "" {
		orderState = sys.CurrentOrders[f][b].Status
	}
	switch orderState {
	case et.Received:
		return "R"
	case et.Accepted:
		if (*sys).ID == sys.CurrentOrders[f][b].Assignee {
			return "A"
		}
		return "a"
	case et.Finished:
		return "F"
	default:
		if b == 0 && f == 3 {
			return " "
		} else if b == 1 && f == 0 {
			return " "
		}
		return "-"
	}
}

//exitPoller signals to outChan when the character p is read.
func exitPoller(outChan chan<- bool) {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if rune(b[0]) == 'p' {
			select {
			case outChan <- true:
				return
			default:
			}

		}
	}
}
