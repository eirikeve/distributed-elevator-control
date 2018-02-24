
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
Run `go get https://github.com/sirupsen/logrus` and the `go install` in src folder to install logging tools.


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
