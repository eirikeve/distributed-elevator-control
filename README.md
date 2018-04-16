# TTK4145 Real Time Programming
_Elevator Project_  - `Group ????` - Spring 2018  
- Person 1 [`github username`]
- Person 2 [`github username`]

## The project

Note: This is the delivery version, where names/info is redacted.

This repository contains our semester project for the course TTK4145 Real Time Programming.  
The task was to make a system capable of real time, robust control and order delegation between N elevators, each with M floors.  
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
    * `main.go`: Starts the system  
* `./sim`, source code and binaries for elevator simulator  
* `./display` source code for our system heads up display, which shows the state & orders of all active systems  



## External Libraries

We have used several external libraries in this project:  
* `logrus`, used for structured information logging to bash with different importance levels  (must be installed from [`www.github.com/sirupsen/logrus`](www.github.com/sirupsen/logrus))
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

The file orderlogic.go in elevfsm, and the orderdelegation module are based on elev_algo and cost_fns in [`https://github.com/TTK4145/Project-resources`](https://github.com/TTK4145/Project-resources).

Our display module has an additional dependency, which is not needed for running the main project:  
* `goterm`, used for making a command line GUI heads up display of the whole elevator system state (must be installed from [`www.github.com/buger/goterm`](www.github.com/buger/goterm))


## Compiling and running

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


## Display

Display is completely separate from the main project.  
It monitors for messages, and uses them to display the active systems states in a HUD.  
Compile by running `go build -o display` in the `./display` folder.
It takes no passed arguments.
