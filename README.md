# TTK4145 Real Time Programming
_Elevator Project_  - `Group 28` - Spring 2018  
- Ole Petter Nordanger [`olepno`](github.com/olepno)
- Eirik Vesterkjær [`eirikeve`](github.com/eirikeve)


## Running the project

You can run the project either with the elevator models in Sanntidssalen on NTNU Gløshaugen, or you can simulate it with the ElevatorSimulator.

### Setting up the ElevatorServer
Call `ElevatorServer` in bash on the computers at Sanntidssalen.
If ElevatorServer command is missing, clone https://github.com/TTK4145/elevator-server and install as described.
Remember to add the folder where the executable is placed to path.

### Setting up the Elevator Simulator
Go to the [`./sim/`](./sim) folder, and run the executable from terminal.

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
Run `go get github.com/sirupsen/logrus` in src folder to install logging tools.


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
