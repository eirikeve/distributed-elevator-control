package sysstate

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	locIP "../elevnetwork/localip"
	et "../elevtype"
	tm "github.com/buger/goterm"
)

var testLocalIP, _ = locIP.LocalIP()
var testLocalID int32 = 32
var testRemoteID int32 = 99

func TestOrderMerge(t *testing.T) {
	//===============================
	// Setup of systems

	splitIP := strings.Split(testLocalIP, ".")
	// Get last byte of the IP
	v, _ := strconv.Atoi(splitIP[len(splitIP)-1])
	testLocalID = int32(v)
	if testLocalID == 0 {
		testLocalID = 256
	}

	var localSysstates []et.ElevState
	localSys := getLocalSys()
	remoteSys := getRemoteSys()

	//===============================
	// Setup of orders
	// Order1 is UP from floor 1. This is the order accepted by Remote
	order1 := getOrder1()
	// Order1 is UP from floor 1. This is the order finished by Local
	order1finished := getOrder1()
	order1finished.Status = et.Finished

	// Order2 is UP from floor 0, received by remote, not transmitted to local yet
	order2 := getOrder2()

	// Order3 is CAB in floor 3 for remote. Accepted
	order3 := getOrder3()
	// Order4 is CAB in floor3 for local. Accepted
	order4 := getOrder3()
	order4.Id = "order4"
	var acks []int32
	acks = append(acks, testLocalID)
	order4.Acks = acks
	order4.Assignee = testLocalID

	// Order5 is DOWN in floor 2. It is acked buy both, finished by remote, but only accepted by local
	order5 := getOrder5()
	order5fin := getOrder5finished()

	//===============================
	// Assignment of orders

	localSys.CurrentOrders[order4.Order.Floor][int(order4.Order.Button)] = order4
	localSys.CurrentOrders[order5.Order.Floor][int(order5.Order.Button)] = order5
	localSys.FinishedOrders[0] = order1finished

	remoteSys.CurrentOrders[order1.Order.Floor][int(order1.Order.Button)] = order1
	remoteSys.CurrentOrders[order2.Order.Floor][int(order2.Order.Button)] = order2
	remoteSys.CurrentOrders[order3.Order.Floor][int(order3.Order.Button)] = order3
	remoteSys.FinishedOrders[0] = order5fin

	//===============================
	// Setup of local sysstate
	localSysstates = append(localSysstates, localSys)

	SetSystemsStates(localSysstates)

	//===============================
	// Tests

	println("Before update: ")
	println(testRemoteID, "remote")
	printSystemInfoAsTable(remoteSys)
	println("-")
	println(testLocalID, testLocalIP)
	printSystemInfoAsTable(GetSystemsStates()[0])

	HandleRegularUpdate(remoteSys)

	println("After update: ")
	println(testLocalID, testLocalIP)
	printSystemInfoAsTable(GetSystemsStates()[0])
}

func getLocalSys() et.ElevState {
	var localSys et.ElevState
	localSys.ID = testLocalID
	localSys.StartupTime = time.Now().Add(-time.Second).Unix()
	localSys.E.Floor = 1
	localSys.E.MovementDirection = et.MD_Stop
	localSys.E.MovDirFromLastFloor = et.MD_Down
	localSys.E.State = et.Idle
	localSys.E.ErrorState = et.FullFunctionality
	return localSys
}

func getRemoteSys() et.ElevState {
	var remoteSys et.ElevState

	remoteSys.ID = testRemoteID
	remoteSys.StartupTime = time.Now().Add(-time.Second * 2).Unix()
	remoteSys.E.Floor = 2
	remoteSys.E.MovementDirection = et.MD_Stop
	remoteSys.E.MovDirFromLastFloor = et.MD_Down
	remoteSys.E.State = et.Idle
	remoteSys.E.ErrorState = et.FullFunctionality

	return remoteSys
}

func getOrder1() et.ElevOrder {
	order1 := et.EmptyOrder()
	order1.Id = "order1"
	order1.Order.Floor = 1
	order1.Order.Button = et.BT_HallUp
	order1.Status = et.Accepted
	order1.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 100).Unix()
	order1.Assignee = testLocalID
	order1.Acks = append(order1.Acks, testLocalID)
	order1.Acks = append(order1.Acks, testRemoteID)

	return order1
}

func getOrder2() et.ElevOrder {
	order2 := et.EmptyOrder()
	order2.Id = "order2"
	order2.Order.Floor = 0
	order2.Order.Button = et.BT_HallUp
	order2.Status = et.Received
	order2.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 180).Unix()
	order2.Assignee = testLocalID
	order2.Acks = append(order2.Acks, testRemoteID)
	return order2
}

func getOrder3() et.ElevOrder {
	order3 := et.EmptyOrder()
	order3.Id = "order4"
	order3.Order.Floor = 3
	order3.Order.Button = et.BT_Cab
	order3.Status = et.Accepted
	order3.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 250).Unix()
	order3.Assignee = testLocalID
	order3.Acks = append(order3.Acks, testLocalID)
	return order3
}

func getOrder4() et.ElevOrder {
	order3 := et.EmptyOrder()
	order3.Id = "order3"
	order3.Order.Floor = 3
	order3.Order.Button = et.BT_Cab
	order3.Status = et.Accepted
	order3.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 200).Unix()
	order3.Assignee = testRemoteID
	order3.Acks = append(order3.Acks, testRemoteID)
	return order3
}

func getOrder5() et.ElevOrder {
	order2 := et.EmptyOrder()
	order2.Id = "order5"
	order2.Order.Floor = 2
	order2.Order.Button = et.BT_HallDown
	order2.Status = et.Accepted
	order2.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 200).Unix()
	order2.Assignee = testRemoteID
	order2.Acks = append(order2.Acks, testLocalID)
	order2.Acks = append(order2.Acks, testRemoteID)
	return order2
}
func getOrder5finished() et.ElevOrder {
	order2 := et.EmptyOrder()
	order2.Id = "order5"
	order2.Order.Floor = 2
	order2.Order.Button = et.BT_HallDown
	order2.Status = et.Finished
	order2.TimestampLastOrderStatusChange = time.Now().Add(-time.Millisecond * 200).Unix()
	order2.Assignee = testRemoteID
	order2.Acks = append(order2.Acks, testLocalID)
	order2.Acks = append(order2.Acks, testRemoteID)
	return order2
}

func printSystemInfoAsTable(sys et.ElevState) {
	minWidth := 0
	tabWidth := 10
	padding := 0
	padChar := byte(' ')
	table := tm.NewTable(minWidth, tabWidth, padding, padChar, 0)

	fmt.Fprintf(table, "State\t \t|\tOrders\t \t \tDone:\t  "+intToBufferedString(int64(len(sys.FinishedOrders)), 4)+"\t|\n")
	fmt.Fprintf(table, "---------\t-----\t|\t---------\t-----\t-----\t-----\t-----\t|\n")
	fmt.Fprintf(table, "Floor \t%d\t|\tFloor\t0\t1\t2\t3\t|\n", sys.E.Floor)
	fmt.Fprintf(table, "State \t%d\t|\tHallUp\t"+orderToString(&sys, 0, 0)+"\t"+orderToString(&sys, 1, 0)+"\t"+orderToString(&sys, 2, 0)+"\t"+orderToString(&sys, 3, 0)+"\t|\n", sys.E.State)
	fmt.Fprintf(table, "ErrState \t%d\t|\tHallDown \t"+orderToString(&sys, 0, 1)+"\t"+orderToString(&sys, 1, 1)+"\t"+orderToString(&sys, 2, 1)+"\t"+orderToString(&sys, 3, 1)+"\t|\n", sys.E.ErrorState)
	fmt.Fprintf(table, "MovDir \t%d\t|\tCab\t"+orderToString(&sys, 0, 2)+"\t"+orderToString(&sys, 1, 2)+"\t"+orderToString(&sys, 2, 2)+"\t"+orderToString(&sys, 3, 2)+"\t|\n", sys.E.MovementDirection)
	print(table.String())

}

func orderToString(sys *et.ElevState, f int, b int) string {
	orderState := et.Unknown
	if sys.CurrentOrders[f][b].Id != "" {
		orderState = sys.CurrentOrders[f][b].Status
	}
	switch orderState {
	case et.Timeout:
		return "T"
	case et.Received:
		return "R"
	case et.Accepted:
		return "A"
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

func intToBufferedString(i int64, sz int) string {
	s := strconv.FormatInt(i, 10)
	if len(s) < sz {
		for i := 0; i < (sz - len(s)); i++ {
			s = " " + s
		}
	}
	return s
}
