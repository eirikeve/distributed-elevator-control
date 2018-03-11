package sysbackup

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

//var backuplogger log.Logger
var initialized = false
var logFile *os.File
var backupRegexp, _ = regexp.Compile("^backup_[0-9]+.elevlog$")
var idRegexp, _ = regexp.Compile("id=.+ backup=")
var stateRegexp, _ = regexp.Compile("backup={.+}\n$")

func setupSysBackup() {

	filename := "backup_" + strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
	var err error
	logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		// Could not open file
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("sysbackup Setup: Could not open log output file. Defaulting to bash output.")
		initialized = false
	} else {
		initialized = true
	}
}

func Backup(states []et.ElevState) {
	if !initialized {
		setupSysBackup()
	}
	print((*logFile).Name())
	for i := 0; i < len(states); i++ {
		backupElevState(states[i])
	}

}

func Recover(timeLimit time.Time) ([]et.ElevState, error) {

	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.WithField("err", err.Error()).Error("sysbackup Recover: Failed")
		return make([]et.ElevState, 0), err
	}
	backupFilesIndexes := getBackupFileIndexes(files)
	//log.WithField("Match indexes", backupFilesIndexes).Info("Indexes")

	useableBackupIndexes := make([]int, 0)
	for _, index := range backupFilesIndexes {
		if files[index].ModTime().Sub(timeLimit) > 0*time.Second {
			useableBackupIndexes = append(useableBackupIndexes, index)
		}
	}
	sortedIndexes, _ := getBackupFileIndexesSortedInIncreasingTime(files, useableBackupIndexes)
	//log.WithField("Match sorted", sortedIndexes).Info("Indexes")
	//for _, index := range sortedIndexes {
	//log.WithField("File", files[index].Name()).Info("Sorted file order")
	//}

	states := make([]et.ElevState, 0)
	for _, backupIndex := range sortedIndexes {
		applyBackupFromFile(&states, files[backupIndex])
	}

	return states, nil
}

func getBackupFileIndexes(files []os.FileInfo) []int {
	backupFilesIndexes := make([]int, 0)
	for i, f := range files {
		//log.WithField("filename", f.Name()).Info("sysbackup Recover: Found")
		if backupRegexp.MatchString(f.Name()) {
			//log.Info("Name matches regexp")
			backupFilesIndexes = append(backupFilesIndexes, i)
		}
	}
	return backupFilesIndexes
}

func getBackupFileIndexesSortedInIncreasingTime(files []os.FileInfo, backupIndexes []int) ([]int, error) {
	timestamps := make([]int64, len(backupIndexes))
	sortedIndexes := make([]int, len(backupIndexes))
	//println("1")
	for i, backupIndex := range backupIndexes {
		//print("i:" + strconv.FormatInt(int64(backupIndex), 10))
		timestampAsString := strings.TrimSuffix(strings.TrimPrefix(files[backupIndex].Name(), "backup_"), ".elevlog")
		var err error
		timestamps[i], err = strconv.ParseInt(timestampAsString, 10, 64)
		if err != nil {
			return make([]int, 0), err
		}
	}
	//println("2")
	for i, _ := range sortedIndexes {
		minimumIndex := 0
		for j, val := range timestamps {
			if val < timestamps[minimumIndex] {
				minimumIndex = j
			}
		}
		sortedIndexes[i] = backupIndexes[minimumIndex]

		if minimumIndex+1 < len(backupIndexes) {
			backupIndexes = append(backupIndexes[:minimumIndex], backupIndexes[minimumIndex+1:]...)
			timestamps = append(timestamps[:minimumIndex], timestamps[minimumIndex+1:]...)
		} else {
			backupIndexes = backupIndexes[:minimumIndex]
			timestamps = timestamps[:minimumIndex]
		}

	}
	return sortedIndexes, nil
}

func applyBackupFromFile(states *[]et.ElevState, backupFile os.FileInfo) {
	log.WithField("filaname", backupFile.Name()).Debug("sysbackup apply: Applying backup")
	file, err := os.OpenFile(backupFile.Name(), os.O_RDONLY, 0755)
	if err != nil {
		return
	}
	var backupReader = bufio.NewReader(file)
	for {
		line, err := backupReader.ReadString('\n')
		if err != nil {
			return
		}
		//elevatorId := getIDFromBackup(&line)
		elevatorJson := getStateJSONFromBackup(&line)
		//log.WithField("id", elevatorId).Info("ReadLine")
		//log.WithField("json", elevatorJson).Info("ReadLine")
		var state et.ElevState
		jsonErr := json.Unmarshal([]byte(elevatorJson), &state)
		if jsonErr != nil {
			//@BUG this always logs
			//log.WithField("err", jsonErr.Error()).Warn("sysbackup apply: Error applying backup")
		}
		//log.WithFields(log.Fields{
		//	"ID":         elevatorId,
		//	"Floor":      state.E.Floor,
		//	"LastUpdate": state.LastUpdate,
		//}).Info("sysbackup apply: Succesfully Unmarshalled")
		if len(state.ID) > 0 {
			if len(*states) == 0 {
				*states = append(*states, state)
			} else {
				for i, s := range *states {
					if s.ID == state.ID {
						(*states)[i] = state
						break // for i, s := ....
					} else if i == len(*states)-1 {
						// We don't know that ID - might have been lost due to crash etc.
						*states = append(*states, state)
						break // for i, s := ....
					}
				}
			}
			//log.WithField("states", *states).Info("sysbackup apply: States")
		}

		/*for _, state := range *states {
			if state.ID == elevatorId {

			}
		}*/

	}

}

func getStateJSONFromBackup(line *string) string {
	stateWithSuffixAndPrefix := stateRegexp.FindString(*line)
	state := strings.TrimPrefix(strings.TrimSuffix(stateWithSuffixAndPrefix, "\n"), "backup=")
	return state
}

func getIDFromBackup(line *string) string {
	idWithSuffixAndPrefix := idRegexp.FindString(*line)
	id := strings.TrimSuffix(strings.TrimPrefix(idWithSuffixAndPrefix, "id="), " backup=")
	return id
}
func backupElevState(state et.ElevState) {
	buf, _ := json.Marshal(state)
	backup := string(buf)
	backupMsg := "time=" + strconv.FormatInt(time.Now().Unix(), 10) + " id=" + state.ID + " backup=" + backup + "\n"

	logFile.Write([]byte(backupMsg))
}

func backupTestWrite(str string) {
	logFile.Write([]byte(str))
}

/*for {
	n, _, _ := conn.ReadFrom(buf[0:])
	for _, ch := range chans {
		T := reflect.TypeOf(ch).Elem()
		typeName := T.String()
		if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
			v := reflect.New(T)
			json.Unmarshal(buf[len(typeName):n], v.Interface())

			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(ch),
				Send: reflect.Indirect(v),
			}})
		}
	}
}*/
