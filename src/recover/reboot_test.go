package recover

import (
	"testing"
	"fmt"
	"time"
	"bufio"
	"os"
	"strings"
)

/* Test created to check if the Secondary is able to monitor the Primary
 * If Primary crashes, Secondary will become after 3 seconds and create a new
 * Secondary.
 * To run test spawnbackup() must be modified to execute "go test"  
 */
func TestSurveillanceProcess(*testing.T){

runSurveillanceProcess()
for {
	//Do nothing
	}
}

func TestReboot(*testing.T){

	userInput := make(chan bool)
	go readUserInput(userInput)
	println("Testing Reboot")
	crashVariable := false


	for {
		select{
		case msg:= <-userInput:
			crashVariable = msg
		default:	
			//println("Running program loop")
			if crashVariable{
				Reboot()
				break;
			}
		}
}
println("Exiting Program loop")
}
var port string
func readUserInput(signalCrash chan bool) {

reader := bufio.NewReader(os.Stdin)
fmt.Print("Enter text: ")
port, _ = reader.ReadString('\n')
port = strings.Replace(port, "\n", "", -1)

for {
	if port != ""{
		time.Sleep(time.Second*5)
		signalCrash <- true
	}
}