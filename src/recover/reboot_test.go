package recover

import (
	"testing"
	"fmt"
	"time"
)

func TestReboot(*testing.T){
	fmt.Printf("Starting reboot test \n\n")
	time.Sleep(time.Second*3)
	Reboot()
}