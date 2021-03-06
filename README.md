# TTK4145 Real Time Programming
_Elevator Project_  - `Group 28` - Spring 2018  
- Ole Petter Nordanger [`olepno`](github.com/olepno)
- Eirik Vesterkjær [`eirikeve`](github.com/eirikeve)

## The project

This repository contains our semester project for the course TTK4145 Real Time Programming.  
The task was to make a distributed system capable of real time, robust control and order delegation between M elevators, each with N floors - running on M different computers.  

We implemented our solution using Golang, with the elevators communicating over UDP. The system has a P2P structure and uses a order acceptance & delegation protocol intended to ensure consistency and order completion.  

Our project has the following parts:  
* `./src`, source code, consisting of the following modules:  
    * `elevdriver`: Basic i/o interaction with the ElevatorServer/simulator. "Hardware interface."  
    * `elevfsm`: Local system state machine & methods for determining when to move/stop, and which orders to clear when.  
    * `elevhandler`: "main-loop" of the local elevator. Sends messages to elevdriver and nethandler through chans, and informs elevfsm of events.  
    * `elevnetwork`: Module with network-related code. Broadcasting, getting local IP address, heartbeat system, etc.  
    * `elevorderevaluation`: Used by nethandler for determining which system to delegate a received order to.  
    * `elevtimer`: Timer module used in elevfsm for signalling timeouts after unloading/initializing etc.  
    * `elevtype`: Definitions of structs and types.  
    * `nethandler`: "main-loop" of the communications/delegation part of the system. Sends order queue and lights from sysstate to elevhandler,  sends/recvs messages from other systems, receives and delegates orders, updates sysstate with the state of the system, regularly backs up the system, etc.  
    * `phoenix`: Module which lets us restart the system if it shuts down unexpectedly.  
    * `setup`: Setup performed at startup  
    * `sysbackup`: Lets us back up the system to a file and recover it.  
    * `sysstate`: Has the state of all active systems. Has logic which lets us incorporate changes in other systems (such as new orders or timeouts in orders) into our system. All logic for accepting/rejecting/acknowledging orders is here.  
    * `main.go`: Main loop  
* `./sim`, source code and binaries for elevator simulator  
* `./display` source code for our system heads up display, which shows the state & orders of all active systems  



## External Libraries

We have used several external libraries in this project:  
* [`logrus`](www.github.com/sirupsen/logrus), used for structured information logging to bash with different importance levels  
* `strconv`, for parsing and formatting numbers  
* `fmt`, for string formatting  
* `net`, for networking in phoenix  
* `os/exec`, for running commands when restarting a system  
* `sync`, for WaitGroup, and also mutex in elevtimer  
* `time`, for getting the current time  
* `encoding/json`, for marshalling/unmarshalling structs to JSON  
* `bufio`, for reading through files for sysbackup  
* `io/ioutil`, for finding files in a directory for sysbackup  
* `os`, for exiting the system in case of failure to initialize, creating a directory for sysbackup (if nonexistens), and opening files  
* `regexp`, for determining which files are our backup files, and finding the information in a line from a backup  
* `strings` for string modification  

And our display module has an additional dependency, which is not needed for running the main project:  
* [`gotem`](www.github.com/buger/goterm), used for making a command line GUI heads up display of the whole elevator system state


## Running the project

You can run the project either with the elevator models in Sanntidssalen on NTNU Gløshaugen, or you can simulate it with the Elevator Simulator.

Running it on the models requires first running ElevatorServer in a terminal.

To run directly, simply call `go run main.go` in `./src`.  
You can also compile with `go build -o elevator` in `./src`.

Available flags for running the elevator (none required):  
* `isDebugEnvironment (bool, default false)`: Toggles whether to show log messages from all log levels (Debug through Panic)
* `doLog (bool, default true)`: Toggles whether to enable logging. Set to false to disable log.  
* `doLogToFile (bool, default false)`: Toggles whether to log to a file in the same folder as the executable, instead of to bash.  
* `ipPort (string, default "15657")`: Port to use for dialing the ElevatorServer/simulator.  
* `backupPort (string, default "23003")`: Port to use for dialing backup.



### Setting up the ElevatorServer
Call `ElevatorServer` in bash on the computers at Sanntidssalen.
If ElevatorServer command is missing, clone https://github.com/TTK4145/elevator-server and install as described.
Remember to add the folder where the executable is placed to path.

### Setting up the Elevator Simulator
Go to the [`./sim/`](./sim) folder, and run the executable from terminal.
If you want to run several elevators locally, set the `--port` flag of each instance differently.


```
+-----------+-----------------+
|           |        #>       |
| Floor     |  0   1*  2   3  |Connected
+-----------+-----------------+-----------+
| Hall Up   |  *   -   -      | Door:   - |
| Hall Down |      -   -   *  | Stop:   - |
| Cab       |  -   -   *   -  | Obstr:  ^ |
+-----------+-----------------+---------43+
```


## Setup of Go project and files



### Programming environment

In bash, run:
```bash
sudo apt-get install code
apt-get install golang-go
cd $HOME
mkdir go
cd go
git clone https://www.github.com/TTK4145/project-eirik-op.git
xdg-open ~/.bashrc
```
Add 
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
``` 
to .bashrc

Reopen terminal, run `go env` to verify GOPATH  



In `VS Code`:  
Add `"go.docsTool": "gogetdoc"` to the User Settings, to avoid the annoying Godoc warning.

__Alternative if__ `apt.get install golang-go` __fails:__ 
In web browser, open [`https://golang.org/dl/`](https://golang.org/dl/), and download go.



### Cleanup after session

Check if all changes are committed:
```bash
cd $GOPATH/project-eirik-op
git diff
```
If needed, commit changes. Beware of untracked files!

Delete the local files and verify:
```
cd $GOPATH
rm -rf project-eirik-op
rm -rf ~/.local/share/Trash/*
ls
```
Log out of everything.
