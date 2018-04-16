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

var initialized = false
var numUseableBackupFiles string
var logFile *os.File
var backupRegexp, _ = regexp.Compile("^backup_[0-9]+.elevlog$")
var idRegexp, _ = regexp.Compile("id=.+ backup=")
var stateRegexp, _ = regexp.Compile("backup={.+}\n$")

const folderDir = "../backup/"

func setupSysBackup() {

	filename := folderDir + "backup_" + strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
	var err error
	logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		// Could not open file
		os.MkdirAll(folderDir, os.ModePerm)
		log.WithError(err).Error("sysbackup Setup: Could not open log output file. Created backup folder " + folderDir)

		logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Error("sysbackup Setup: Could not initialize")
			initialized = false
		} else {
			initialized = true

		}

	} else {
		initialized = true

	}
}

func Backup(states []et.ElevState) {
	if !initialized {
		setupSysBackup()
	}
	log.WithField("FileName", (*logFile).Name()).Debug("sysbackup Backup: Backed up")
	for i := 0; i < len(states); i++ {
		backupElevState(states[i])
	}

}

func Recover(timeLimit time.Time) ([]et.ElevState, error) {

	files, err := ioutil.ReadDir(folderDir)
	if err != nil {
		log.WithField("err", err.Error()).Error("sysbackup Recover: Failed")
		return make([]et.ElevState, 0), err
	}
	backupFilesIndexes := getBackupFileIndexes(files)

	useableBackupIndexes := make([]int, 0)
	for _, index := range backupFilesIndexes {
		if files[index].ModTime().Sub(timeLimit) > 0*time.Second {
			useableBackupIndexes = append(useableBackupIndexes, index)
			log.WithField("usable file", files[index].Name()).Info("sysbackup Recover: Valid file for recovery")
		}
	}

	sortedIndexes, _ := getBackupFileIndexesSortedInIncreasingTime(files, useableBackupIndexes)

	states := make([]et.ElevState, 0)
	for _, backupIndex := range sortedIndexes {
		log.WithField("File", files[backupIndex].Name()).Info("sysbackup: Applying backup")
		applyBackupFromFile(&states, files[backupIndex])
	}
	numBackupFiles := strconv.FormatInt(int64(len(backupFilesIndexes)), 10)
	numUseableBackupFiles = strconv.FormatInt(int64(len(useableBackupIndexes)), 10)
	numStatesRecovered := strconv.FormatInt(int64(len(states)), 10)

	log.WithFields(log.Fields{"numStatesRecovered": numStatesRecovered,
		"numBackupFiles":        numBackupFiles,
		"numUseableBackupFiles": numUseableBackupFiles,
	}).Info("sysbackup Recover:")

	return states, nil
}

func getBackupFileIndexes(files []os.FileInfo) []int {
	backupFilesIndexes := make([]int, 0)
	for i, f := range files {
		if backupRegexp.MatchString(f.Name()) {
			backupFilesIndexes = append(backupFilesIndexes, i)
		}
	}
	return backupFilesIndexes
}

func getBackupFileIndexesSortedInIncreasingTime(files []os.FileInfo, backupIndexes []int) ([]int, error) {
	timestamps := make([]int64, len(backupIndexes))
	sortedIndexes := make([]int, len(backupIndexes))
	for i, backupIndex := range backupIndexes {
		timestampAsString := strings.TrimSuffix(strings.TrimPrefix(files[backupIndex].Name(), "backup_"), ".elevlog")
		var err error
		timestamps[i], err = strconv.ParseInt(timestampAsString, 10, 64)
		if err != nil {
			return make([]int, 0), err
		}
	}
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
	file, err := os.OpenFile(folderDir+backupFile.Name(), os.O_RDONLY, 0755)
	if err != nil {
		log.WithError(err).Error("sysbackup: Unable to apply backup from this file")
		return
	}
	var backupReader = bufio.NewReader(file)
	for {
		line, err := backupReader.ReadString('\n')
		if err != nil {
			return
		}
		elevatorJson := getStateJSONFromBackup(&line)
		var state et.ElevState
		jsonErr := json.Unmarshal([]byte(elevatorJson), &state)
		if jsonErr != nil {
			log.WithField("err", jsonErr.Error()).Warn("sysbackup apply: Error applying backup")
		}
		if state.ID > 0 {
			if len(*states) == 0 {
				*states = append(*states, state)
			} else {
				for i, s := range *states {
					if s.ID == state.ID {
						(*states)[i] = state
						break // for i, s := range *states
					} else if i == len(*states)-1 {
						*states = append(*states, state)
						break // for i, s := range *states
					}
				}
			}
		}
	}
}

func getStateJSONFromBackup(line *string) string {
	stateWithSuffixAndPrefix := stateRegexp.FindString(*line)
	state := strings.TrimPrefix(strings.TrimSuffix(stateWithSuffixAndPrefix, "\n"), "backup=")
	return state
}

func GetIDFromBackup(line *string) string {
	idWithSuffixAndPrefix := idRegexp.FindString(*line)
	id := strings.TrimSuffix(strings.TrimPrefix(idWithSuffixAndPrefix, "id="), " backup=")
	return id
}
func backupElevState(state et.ElevState) {
	if logFile == nil {
		log.WithField("err", "Nonexistent logfile").Error("sysbackup backup: Could not back up")
		return
	}
	buf, _ := json.Marshal(state)
	backup := string(buf)
	backupMsg := "time=" + strconv.FormatInt(time.Now().Unix(), 10) + " id=" + strconv.FormatInt(int64(state.ID), 10) + " backup=" + backup + "\n"

	logFile.Write([]byte(backupMsg))
}
