package elevtype

import (
	"testing"

	"../elevlog"
	log "github.com/sirupsen/logrus"
)

func TestOrder(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)

	log.Info("Running Test on GeneralOrder")
	log.Info("Empty orders")
	var order1 SimpleOrder
	var order2 ElevOrder
	logGeneralOrder(order1, order2)
	log.Info("Only IDs set (different for each order)")
	order1.Id = "Order 1"
	order2.Id = "Order 2"
	logGeneralOrder(order1, order2)
	log.Info("Order set as well (different for each order)")
	order1.Order = ButtonEvent{0, BT_HallUp}
	order2.Order = ButtonEvent{2, BT_HallUp}
	logGeneralOrder(order1, order2)

}

func logGeneralOrder(order1 GeneralOrder, order2 GeneralOrder) {
	log.Info("GeneralOrder printing tests:")
	log.WithField("order1.isSame(order2)", order1.IsSame(order2)).Debug("Checking isSame()")
	log.WithField("order2.isSame(order1)", order2.IsSame(order1)).Debug("Checking isSame()")
	log.WithField("order1.isEmpty()", order1.IsEmpty()).Debug("Checking IsEmpty()")
	log.WithField("order2.isEmpty()", order2.IsEmpty()).Debug("Checking IsEmpty()")
	log.WithField("order1.GetID()", order1.GetID()).Debug("Checking GetID()")
	log.WithField("order2.GetID()", order2.GetID()).Debug("Checking GetID()")
	log.WithField("order1.GetOrder()", order1.GetOrder()).Debug("Checking GetOrder()")
	log.WithField("order2.GetOrder()", order2.GetOrder()).Debug("Checking GetOrder()")
	order1copy := order1.ToSimpleOrder()
	order2copy := order2.ToSimpleOrder()
	log.Debug("Testing copying with ToSimpleOrder")
	log.WithField("order1.isSame(order1copy)", order1.IsSame(order1copy)).Debug("Checking isSame()")
	log.WithField("order2.isSame(order2copy)", order2.IsSame(order2copy)).Debug("Checking isSame()")
}
